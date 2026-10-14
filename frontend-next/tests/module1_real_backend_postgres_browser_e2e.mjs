// Real Next.js  + unmocked HTTP Go API + actual PG18. The CDP connection
// only drives the browser and observes network; NO Fetch interception exists.
import assert from "node:assert/strict";
import {readFileSync,writeFileSync} from "node:fs";
import {setTimeout as sleep} from "node:timers/promises";
const fixture=JSON.parse(readFileSync(process.env.OPENINVEST_AUDIT_PROOF_FILE,"utf8"));
const frontend="http://127.0.0.1:3000";
const observed=[];
const requestMethods=new Map();
class CDP{
 constructor(ws){this.ws=ws;this.id=0;this.waiters=new Map();this.callbacks=new Map();
  ws.onmessage=event=>{const e=JSON.parse(event.data);if(e.id){
   const h=this.waiters.get(e.id);if(h){this.waiters.delete(e.id);e.error?h.reject(Error(JSON.stringify(e.error))):h.resolve(e.result)}
  }else if(e.method){this.callbacks.get(e.method)?.(e.params)}};
 }
 send(method,params={}){const id=++this.id;return new Promise((resolve,reject)=>{
  this.waiters.set(id,{resolve,reject});this.ws.send(JSON.stringify({id,method,params}));
  setTimeout(()=>{if(this.waiters.has(id)){this.waiters.delete(id);reject(Error("CDP timeout "+method))}},20000).unref();
 })}
 on(event,handler){this.callbacks.set(event,handler)}
 close(){this.ws.close()}
}
let cdp;
async function js(expression){
 const r=await cdp.send("Runtime.evaluate",{expression,returnByValue:true,awaitPromise:true,userGesture:true});
 if(r.exceptionDetails)throw Error("CDP JS evaluation exception: "+JSON.stringify(r.exceptionDetails));
 return r.result.value;
}
async function poll(label,test,ms=16000){const start=Date.now();while(Date.now()-start<ms){if(await test())return;await sleep(90)}throw Error("BROWSER_CONDITION_NOT_MET "+label)}
async function visit(token){
 await cdp.send("Page.navigate",{url:frontend+"/auth/verify#"+token});
 await poll("real Next verification form",()=>js("Boolean(document.querySelector('form.form-grid input[type=password]'))"));
 await poll("raw fragment scrubbed",()=>js("location.hash === ''"));
 await poll("verification submit enabled",()=>js("Boolean(document.querySelector('form.form-grid button:not([disabled])'))"));
}
async function submit(){
 await js("document.querySelector('input[type=password]').focus(); true");
 await cdp.send("Input.insertText",{text:"actual-browser-proof-password-2026"});
 await poll("React new-password state",()=>js("document.querySelector('input[type=password]').value.length>15"));
 await js("document.querySelector('form.form-grid button').click();true");
}
async function state(){return js("({ready:!!document.querySelector('a[href=\\\"/\\\"]'),message:document.querySelector('[role=\\\"status\\\"]')?.textContent||'',hash:location.hash,local:localStorage.length,session:sessionStorage.length,cookie:document.cookie})")}
async function main(){
 await poll("headless Chrome CDP endpoint",async()=>{try{return (await fetch("http://127.0.0.1:9222/json/list")).ok}catch{return false}});
 const targets=await(await fetch("http://127.0.0.1:9222/json/list")).json();
 const target=targets.find(x=>x.type==="page"&&x.webSocketDebuggerUrl);assert.ok(target,"no real Chrome page");
 const ws=new WebSocket(target.webSocketDebuggerUrl);
 await new Promise((resolve,reject)=>{ws.onopen=resolve;ws.onerror=reject});
 cdp=new CDP(ws);
 await cdp.send("Page.enable");await cdp.send("Runtime.enable");await cdp.send("Network.enable");
 cdp.on("Network.requestWillBeSent",event=>{
  requestMethods.set(event.requestId,event.request.method);
 });
 cdp.on("Network.responseReceived",event=>{
  if(event.response.url.includes("127.0.0.1:8080/api/v1/auth/email/verify") && requestMethods.get(event.requestId)==="POST")
   observed.push({status:event.response.status,method:"POST"});
 });
 try{
  await visit(fixture.proof);
  let s=await state();assert.equal(s.hash,"");assert.equal(s.local,0);assert.equal(s.session,0);
  await submit();
  await poll("actual PG-backed account activation",async()=>(await state()).ready,24000);
  await poll("HTTP server returned success",()=>observed.length>=1);
  s=await state();
  assert.equal(observed[0].status,200,"unmocked backend verification HTTP must return 200");
  assert.equal(s.ready,true);assert.equal(s.hash,"");assert.equal(s.local,0);assert.equal(s.session,0);
  assert.ok(!s.cookie.includes("refresh"),"activation incorrectly left refresh cookie");
  process.stdout.write("REAL_CHROMIUM_NEXTJS_GO_HTTP_PG_ACTIVATION=PASS HTTP_STATUS=200 TOKEN_FRAGMENT_SCRUBBED=YES NO_BROWSER_TOKEN_STORAGE=YES AUTO_AUTHENTICATED_SESSION=NO\n");
  await visit(fixture.proof);await submit();
  await poll("replayed token refused by real backend",()=>observed.length>=2,24000);
  await sleep(450);
  s=await state();
  assert.ok(observed[1].status>=400&&observed[1].status<500,"used proof did not get 4xx");
  assert.equal(s.ready,false,"reused proof activated account");
  assert.equal(s.local,0);assert.equal(s.session,0);assert.equal(s.hash,"");
  process.stdout.write("REAL_BACKEND_PG_CONSUMED_PROOF_REPLAY=DENIED NO_STORAGE_LEAK=YES NO_AUTO_SESSION=YES\n");
 }catch(error){
  try{const a=await cdp.send("Page.captureScreenshot",{format:"png"});writeFileSync("/tmp/module1-real-browser-failure.png",Buffer.from(a.data,"base64"))}catch{}
  throw error;
 }finally{cdp.close()}
}
main().catch(e=>{console.error("REAL_BROWSER_BACKEND_PG_E2E_FAILED",e.stack||String(e));process.exitCode=1});
