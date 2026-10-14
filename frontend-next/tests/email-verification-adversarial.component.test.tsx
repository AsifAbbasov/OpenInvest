// @ts-nocheck
import assert from "node:assert/strict";
import test from "node:test";
import { JSDOM } from "jsdom";
const dom = new JSDOM("<!doctype html><html><body></body></html>",{url:"https://localhost/auth/verify"});
for(const [name,value] of Object.entries({window:dom.window,document:dom.window.document,navigator:dom.window.navigator,HTMLElement:dom.window.HTMLElement,Event:dom.window.Event,MouseEvent:dom.window.MouseEvent,HTMLInputElement:dom.window.HTMLInputElement})){
 Object.defineProperty(globalThis,name,{configurable:true,writable:true,value});
}
Object.defineProperty(globalThis,"IS_REACT_ACT_ENVIRONMENT",{configurable:true,writable:true,value:true});
const React=await import("react");
const {createRoot}=await import("react-dom/client");
const {default:VerifyEmailPage}=await import("../src/app/auth/verify/page.tsx");
const originalFetch=globalThis.fetch;
function response(status,body){return new Response(JSON.stringify(body),{status,headers:{"Content-Type":"application/json"}})}
async function mount(secret){
 dom.window.history.replaceState(null,"","/auth/verify#"+secret);
 dom.window.document.body.innerHTML='<div id="root"></div>';
 const container=dom.window.document.getElementById("root");const root=createRoot(container);
 await React.act(async()=>root.render(React.createElement(React.StrictMode,null,React.createElement(VerifyEmailPage))));
 assert.equal(dom.window.location.hash,"");assert.equal(dom.window.localStorage.length,0);
 return {container,root};
}
async function inputPassword(container,password){
 const el=container.querySelector('input[type="password"]');
 assert.ok(el);const setter=Object.getOwnPropertyDescriptor(dom.window.HTMLInputElement.prototype,"value").set;
 await React.act(async()=>{setter.call(el,password);el.dispatchEvent(new dom.window.Event("input",{bubbles:true}));});
 assert.equal(el.value,password);
}
async function submitVerification(container){
 const form=container.querySelector("form.form-grid");
 assert.ok(form);
 await React.act(async()=>form.dispatchEvent(new dom.window.Event("submit",{bubbles:true,cancelable:true})));
}
test("mounted verification requires a token, leaves no browser persistent secret",async t=>{
 t.after(()=>{globalThis.fetch=originalFetch});
 let count=0;globalThis.fetch=async()=>{count++;throw Error("no request without token")};
 const {container,root}=await mount("");
 try{
  assert.equal(container.querySelector('button')?.disabled,true);
  assert.equal(dom.window.localStorage.length,0);assert.equal(dom.window.sessionStorage.length,0);
  assert.equal(count,0);
 }finally{await React.act(async()=>root.unmount())}
});
test("mounted verification exposes invalid and expired token errors without authentication",async t=>{
 t.after(()=>{globalThis.fetch=originalFetch});
 const token="T".repeat(43);let count=0,seen;
 globalThis.fetch=async(url,opts)=>{count++;seen={url,opts};return response(400,{error:{code:"INVALID_VERIFICATION",message:"Verification could not be completed. Request new instructions."},meta:{requestId:"synthetic"}})};
 const {container,root}=await mount(token);
 try{
  await inputPassword(container,"owner-selected-password-2026");
  await submitVerification(container);
  assert.equal(count,1);
  assert.equal(seen.opts.credentials,"omit");
  assert.equal(seen.opts.method,"POST");
  assert.ok(seen.url.endsWith("/api/v1/auth/email/verify"));
  assert.equal(JSON.parse(seen.opts.body).token,token);
  assert.match(container.querySelector('[role="status"]').textContent,/Verification could not be completed/);
  assert.equal(container.querySelector("a[href='/']"),null);
  assert.equal(dom.window.localStorage.length,0);assert.equal(dom.window.sessionStorage.length,0);
 }finally{await React.act(async()=>root.unmount())}
});
test("mounted verification shows sanitized network error, permits retry without leaking token",async t=>{
 t.after(()=>{globalThis.fetch=originalFetch});
 const token="N".repeat(43);let count=0;
 globalThis.fetch=async(url,opts)=>{count++;if(count===1)throw Error("SENSITIVE_REMOTE_NETWORK_DETAIL");return response(200,{data:{status:"ACCOUNT_READY"},meta:{requestId:"synthetic"}})};
 const {container,root}=await mount(token);
 try{
  await inputPassword(container,"owner-selected-password-2026");
  await submitVerification(container);
  assert.equal(count,1);
  assert.match(container.querySelector('[role="status"]').textContent,/temporarily unavailable/);
  assert.doesNotMatch(container.textContent,/SENSITIVE_REMOTE_NETWORK_DETAIL/);
  await submitVerification(container);
  assert.equal(count,2);
  assert.ok(container.querySelector("a[href='/']"));
  assert.equal(dom.window.location.hash,"");
  assert.equal(dom.window.localStorage.length,0);
 }finally{await React.act(async()=>root.unmount())}
});


test("ATTACK: changing verification link while page remains mounted must replace A token with B",async t=>{
 t.after(()=>{globalThis.fetch=originalFetch});
 const a="A".repeat(43),b="B".repeat(43);const requests=[];
 globalThis.fetch=async(url,opts)=>{requests.push(JSON.parse(opts.body));return response(400,{error:{code:"INVALID_VERIFICATION",message:"Invalid token."}})};
 const {container,root}=await mount(a);
 try{
  await inputPassword(container,"owner-selected-password-2026");
  dom.window.history.replaceState(null,"","/auth/verify#"+b);
  await React.act(async()=>{dom.window.dispatchEvent(new dom.window.Event("hashchange"));});
  await submitVerification(container);
  assert.equal(requests.length,1,"negative control: a verification request must be sent");
  assert.equal(requests[0].token,b,"STALE_PRINCIPAL: token A was submitted after navigation to B");
  assert.equal(dom.window.location.hash,"","new verification secret must be removed from URL");
 }finally{await React.act(async()=>root.unmount())}
});
test("ATTACK: old verification response must not mark a different link/account ready",async t=>{
 t.after(()=>{globalThis.fetch=originalFetch});
 const a="C".repeat(43),b="D".repeat(43);let release,received=null;
 globalThis.fetch=async(url,opts)=>{received=JSON.parse(opts.body).token;return new Promise(resolve=>{release=()=>resolve(response(200,{data:{status:"ACCOUNT_READY"}}));});};
 const {container,root}=await mount(a);
 try{
  await inputPassword(container,"owner-selected-password-2026");
  const form=container.querySelector("form.form-grid");assert.ok(form);
  await React.act(async()=>{form.dispatchEvent(new dom.window.Event("submit",{bubbles:true,cancelable:true}));});
  assert.equal(received,a,"negative control: old principal request was never sent");
  assert.equal(typeof release,"function");
  dom.window.history.replaceState(null,"","/auth/verify#"+b);
  await React.act(async()=>{dom.window.dispatchEvent(new dom.window.Event("hashchange"));});
  await React.act(async()=>{release();await new Promise(resolve=>setTimeout(resolve,1));});
  assert.equal(container.querySelector("a[href='/']"),null,"STALE_RESPONSE: old A request marked the new B context ready");
  assert.doesNotMatch(container.querySelector('[role="status"]').textContent,/Account ready/);
 }finally{await React.act(async()=>root.unmount())}
});

test("NEGATIVE CONTROL: deferred success without principal switch must show account ready",async t=>{
 t.after(()=>{globalThis.fetch=originalFetch});
 const old="E".repeat(43);let release,seen;
 globalThis.fetch=async(url,opts)=>{seen=JSON.parse(opts.body).token;return new Promise(resolve=>{release=()=>resolve(response(200,{data:{status:"ACCOUNT_READY"}}));});};
 const {container,root}=await mount(old);
 try{
  await inputPassword(container,"owner-selected-password-2026");
  const form=container.querySelector("form.form-grid");assert.ok(form);
  await React.act(async()=>{form.dispatchEvent(new dom.window.Event("submit",{bubbles:true,cancelable:true}));});
  assert.equal(seen,old);assert.equal(typeof release,"function");
  await React.act(async()=>{release();await new Promise(resolve=>setTimeout(resolve,1));});
  assert.ok(container.querySelector("a[href='/']"),"HARNESS_ORACLE_DEFECT: deferred success never reaches mounted page");
 }finally{await React.act(async()=>root.unmount())}
});
