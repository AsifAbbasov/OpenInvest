// Audit-only: real Chromium, built Next.js, live Go HTTP API, disposable PostgreSQL.
// No request interception, mock response, or access to external services.
import assert from "node:assert/strict";
import fs from "node:fs";
import {setTimeout as sleep} from "node:timers/promises";
const fixture=JSON.parse(fs.readFileSync(process.env.AUTHGAP_BROWSER_PROOF_FILE,"utf8"));
const invalid="Z".repeat(43);
const wait=async(label,fn)=>{for(let i=0;i<130;i++){if(await fn())return;await sleep(90)}throw Error("NOT_OBSERVED "+label)};
const pages=await (await fetch("http://127.0.0.1:9222/json/list")).json();
const target=pages.find(p=>p.type==="page"&&p.webSocketDebuggerUrl);
assert.ok(target);
const ws=new WebSocket(target.webSocketDebuggerUrl);
await new Promise((ok,no)=>{ws.onopen=ok;ws.onerror=no});
let serial=0;const pending=new Map();
ws.onmessage=evt=>{const value=JSON.parse(evt.data);if(!value.id)return;const p=pending.get(value.id);if(p){pending.delete(value.id);value.error?p.reject(Error(JSON.stringify(value.error))):p.resolve(value.result)}};
const cmd=(method,params={})=>{const id=++serial;return new Promise((resolve,reject)=>{pending.set(id,{resolve,reject});ws.send(JSON.stringify({id,method,params}))})};
const js=async code=>{const response=await cmd("Runtime.evaluate",{expression:code,returnByValue:true,awaitPromise:true,userGesture:true});if(response.exceptionDetails)throw Error(JSON.stringify(response.exceptionDetails));return response.result.value};
const status=()=>js("document.querySelector('[role=status]')?.textContent||''");
const state=()=>js("({hash:location.hash,ready:!!document.querySelector('a[href=\"/\"]'),local:localStorage.length,session:sessionStorage.length})");
await cmd("Page.enable");await cmd("Runtime.enable");
try{
 await cmd("Page.navigate",{url:"http://127.0.0.1:3000/auth/verify#"+invalid});
 await wait("Next.js form",()=>js("!!document.querySelector('input[type=password]')"));
 await wait("bad fragment scrubbed",()=>js("location.hash===''"));
 await js("document.querySelector('input[type=password]').focus()");
 await cmd("Input.insertText",{text:"isolated-real-api-browser-password-2026"});
 await wait("React input",()=>js("document.querySelector('input[type=password]').value.length>20"));
 await js("document.querySelector('form.form-grid button').click();true");
 await wait("invalid token denied by live Go API",async()=>(await status()).includes("Verification could not be completed"));
 const bad=await state();assert.equal(bad.ready,false);assert.equal(bad.local,0);assert.equal(bad.session,0);
 console.log("REAL_CHROME_NEXT_POSTGRES_INVALID_EMAIL_PROOF=REJECTED");
 await js("history.pushState(history.state,'','/auth/verify#"+fixture.token+"');true");
 await wait("real proof fragment scrubbed",()=>js("location.hash===''"));
 await wait("valid real proof form",()=>js("!!document.querySelector('form.form-grid button:not([disabled])')"));
 await js("document.querySelector('form.form-grid button').click();true");
 await wait("account activated by actual API and PostgreSQL",async()=>(await state()).ready);
 const good=await state();assert.equal(good.hash,"");assert.equal(good.local,0);assert.equal(good.session,0);
 console.log("REAL_CHROME_NEXT_GO_API_POSTGRES_ACCOUNT_READY=PASS NO_AUTOMATIC_BROWSER_SESSION=YES");
 await js("history.back();true");await sleep(350);
 assert.equal((await state()).hash,"");
 await js("history.forward();true");await sleep(350);
 assert.equal((await state()).hash,"");
 const login=await fetch("http://127.0.0.1:8080/api/v1/auth/login",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({email:fixture.email,password:"isolated-real-api-browser-password-2026"})});
 assert.equal(login.status,200);
 console.log("REAL_BACKEND_LOGIN_AFTER_EMAIL_OWNERSHIP=PASS HISTORY_BACK_FORWARD_SCRUB=PASS");
}catch(error){
 try{const shot=await cmd("Page.captureScreenshot",{format:"png"});fs.writeFileSync("/tmp/authgap-fullstack-browser-failure.png",Buffer.from(shot.data,"base64"))}catch{}
 throw error;
}finally{ws.close()}
