// Independent real Chromium + built Next.js App Router test (no JSDOM).
// Uses Chrome DevTools Protocol via Node's built-in WebSocket and a controlled
// browser-network fixture. Never uses real credentials or external SMTP.
import assert from "node:assert/strict";
import { setTimeout as delay } from "node:timers/promises";
import { writeFileSync } from "node:fs";

const base="http://127.0.0.1:3000";
const token=letter=>letter.repeat(43);
const A=token("A"),B=token("B"),C=token("C"),D=token("D");
const events=[];
const pending=new Map();
let cdp;
class Client {
 constructor(socket){this.socket=socket;this.seq=0;this.pending=new Map();this.callbacks=new Map();
 socket.onmessage=event=>{
  const data=JSON.parse(event.data);
  if(data.id){const p=this.pending.get(data.id);if(p){this.pending.delete(data.id);data.error?p.reject(new Error(JSON.stringify(data.error))):p.resolve(data.result)}}
  else if(data.method){const handler=this.callbacks.get(data.method);if(handler)handler(data.params)}
 };
 }
 on(name,fn){this.callbacks.set(name,fn)}
 send(method,params={}){const id=++this.seq;return new Promise((resolve,reject)=>{
  this.pending.set(id,{resolve,reject});this.socket.send(JSON.stringify({id,method,params}));
  setTimeout(()=>{if(this.pending.has(id)){this.pending.delete(id);reject(new Error("CDP timeout "+method))}},12000).unref();
 })}
 close(){this.socket.close()}
}
async function poll(name,fn,limit=10000){const start=Date.now();while(Date.now()-start<limit){if(await fn())return;await delay(90)}throw new Error("BROWSER_NOT_OBSERVED: "+name)}
async function js(expression){const r=await cdp.send("Runtime.evaluate",{expression,returnByValue:true,awaitPromise:true,userGesture:true});if(r.exceptionDetails)throw new Error("JS eval exception: "+JSON.stringify(r.exceptionDetails));return r.result.value}
async function open(url){await cdp.send("Page.navigate",{url});await poll("page rendered",()=>js("Boolean(document.querySelector('input[type=password]'))"),13000);await poll("token scrubbed",()=>js("window.location.hash === ''"),5000)}
async function navigate(letter,method){
 await js("(function(){window.history."+method+"(window.history.state,'','/auth/verify?view=campaign#"+letter+"'); return true})()");
 await poll("router token fragment scrubbed",()=>js("window.location.hash === ''"),2500);
 await poll("verification button enabled",()=>js("Boolean(document.querySelector('form.form-grid button:not([disabled])'))"),2500);
}
async function enterPassword(){
 await js("document.querySelector('input[type=password]').focus()");
 await cdp.send("Input.insertText",{text:"independent-campaign-safe-password-2026"});
 await poll("React password input updated",()=>js("document.querySelector('input[type=password]').value.length > 12"));
}
async function submit(){await js("document.querySelector('form.form-grid button').click(); true")}
async function state(){return js("({ready:Boolean(document.querySelector('a[href=\"/\"]')),message:document.querySelector('[role=\"status\"]')?.textContent||'',hash:location.hash,local:localStorage.length,session:sessionStorage.length})")}
async function fulfill(id,ok){
 const payload=ok?{data:{status:"ACCOUNT_READY"},meta:{requestId:"campaign-test"}}:{error:{code:"INVALID_VERIFICATION",message:"Verification could not be completed. Request new instructions."},meta:{requestId:"campaign-test"}};
 await cdp.send("Fetch.fulfillRequest",{requestId:id,responseCode:ok?200:400,responseHeaders:[
  {name:"Content-Type",value:"application/json"},
  {name:"Access-Control-Allow-Origin",value:base},
  {name:"Access-Control-Allow-Credentials",value:"true"},
  {name:"Cache-Control",value:"no-store"},
 ],body:Buffer.from(JSON.stringify(payload)).toString("base64")});
}
async function main(){
 await poll("Chrome debugging endpoint",async()=>{try{const r=await fetch("http://127.0.0.1:9222/json/list");return r.ok}catch{return false}},12000);
 const targets=await(await fetch("http://127.0.0.1:9222/json/list")).json();
 const page=targets.find(x=>x.type==="page"&&x.webSocketDebuggerUrl);assert.ok(page,"real Chrome page target unavailable");
 const ws=new WebSocket(page.webSocketDebuggerUrl);
 await new Promise((resolve,reject)=>{ws.onopen=resolve;ws.onerror=reject});
 cdp=new Client(ws);
 await cdp.send("Page.enable");await cdp.send("Runtime.enable");await cdp.send("Network.enable");
 cdp.on("Fetch.requestPaused",async p=>{
  try{
   const req=p.request;
   if(req.method==="OPTIONS"){
    await cdp.send("Fetch.fulfillRequest",{requestId:p.requestId,responseCode:204,responseHeaders:[
      {name:"Access-Control-Allow-Origin",value:base},
      {name:"Access-Control-Allow-Credentials",value:"true"},
      {name:"Access-Control-Allow-Methods",value:"POST,OPTIONS"},
      {name:"Access-Control-Allow-Headers",value:"Content-Type,Accept"},
      {name:"Access-Control-Max-Age",value:"0"},
    ]});return;
   }
   if(!req.url.includes("/api/v1/auth/email/verify")){await cdp.send("Fetch.continueRequest",{requestId:p.requestId});return}
   const proof=JSON.parse(req.postData||"{}").token;
   const label=proof===A?"A":proof===B?"B":proof===C?"C":proof===D?"D":"OTHER";
   events.push(label);
   if(label==="A"||label==="C"){pending.set(label,p.requestId);return}
   await fulfill(p.requestId,label==="D");
  }catch(err){process.stderr.write("MOCK_FIXTURE_ERROR: "+String(err)+"\n");process.exitCode=1}
 });
 await cdp.send("Fetch.enable",{patterns:[{urlPattern:"*api/v1/auth/email/verify*",requestStage:"Request"}]});
 try{
  await open(base+"/auth/verify#"+A);
  const original=await state();assert.equal(original.hash,"");assert.equal(original.ready,false);
  await enterPassword();await submit();
  await poll("A request in flight",()=>events.includes("A"));
  assert.ok(pending.has("A"),"A request not actually held");
  await navigate(B,"pushState");
  await submit();
  await poll("B reaches the real HTTP request",()=>events.includes("B"));
  await poll("B error rendered",async()=>((await state()).message.includes("Verification could not be completed")));
  await fulfill(pending.get("A"),true);pending.delete("A");
  await delay(500);
  let current=await state();
  assert.equal(current.ready,false,"STALE_A_SUCCESS_PROMOTED_B");
  assert.ok(current.message.includes("Verification could not be completed"),"STALE_A_SUCCESS_OVERWROTE_B");
  assert.equal(current.local,0);assert.equal(current.session,0);
  process.stdout.write("BROWSER_STALE_A_SUCCESS=REJECTED REAL_NEXTJS=YES REAL_CHROMIUM=YES\n");

  await navigate(C,"replaceState");
  await submit();await poll("C request in flight",()=>events.includes("C"));
  assert.ok(pending.has("C"));
  await navigate(D,"pushState");
  await submit();await poll("D reaches real HTTP request",()=>events.includes("D"));
  await poll("D success rendered",async()=>((await state()).ready));
  await fulfill(pending.get("C"),false);pending.delete("C");
  await delay(500);
  current=await state();
  assert.equal(current.ready,true,"STALE_C_FAILURE_RESET_D");
  assert.equal(current.hash,"");assert.equal(current.local,0);assert.equal(current.session,0);
  process.stdout.write("BROWSER_STALE_C_FAILURE=REJECTED NEXTJS_PUSH_REPLACE=PASS TOKEN_FRAGMENT_SCRUB=PASS NO_BROWSER_TOKEN_STORAGE=PASS NO_AUTO_SESSION=PASS\n");
  process.stdout.write("BROWSER_VERIFICATION_REQUESTS="+events.join(",")+"\n");
 }catch(err){
  try{const image=await cdp.send("Page.captureScreenshot",{format:"png"});writeFileSync("/tmp/openinvest-campaign-browser-failure.png",Buffer.from(image.data,"base64"))}catch{}
  throw err;
 }finally{cdp.close()}
}
main().catch(error=>{console.error("REAL_BROWSER_E2E_FAILED",error.stack||String(error));process.exitCode=1});
