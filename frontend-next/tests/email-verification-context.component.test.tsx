// @ts-nocheck
// Independent bounded adversarial UI probes. Do not replace failure with a mocked
// reducer: these tests mount the production verification page in React StrictMode.
import assert from "node:assert/strict";
import test from "node:test";
import { JSDOM } from "jsdom";
const dom = new JSDOM("<!doctype html><html><body></body></html>",{url:"https://localhost/auth/verify"});
for(const [name,value] of Object.entries({window:dom.window,document:dom.window.document,navigator:dom.window.navigator,HTMLElement:dom.window.HTMLElement,Event:dom.window.Event,HTMLInputElement:dom.window.HTMLInputElement})){
 Object.defineProperty(globalThis,name,{configurable:true,writable:true,value});
}
Object.defineProperty(globalThis,"IS_REACT_ACT_ENVIRONMENT",{configurable:true,writable:true,value:true});
const React=await import("react");
const {createRoot}=await import("react-dom/client");
const {default:VerifyEmailPage}=await import("../src/app/auth/verify/page.tsx");
const originalFetch=globalThis.fetch;
const defer=()=>{let resolve;const promise=new Promise(r=>resolve=r);return {promise,resolve}};
const ok=()=>new Response(JSON.stringify({data:{status:"ACCOUNT_READY"},meta:{requestId:"synthetic-local",traceId:"synthetic-local",generatedAt:"2026-10-14T00:00:00Z"}}),{status:200,headers:{"Content-Type":"application/json"}});
const denied=()=>new Response(JSON.stringify({error:{code:"INVALID_VERIFICATION",message:"Verification could not be completed. Request new instructions."}}),{status:400,headers:{"Content-Type":"application/json"}});
const token=(letter)=>letter.repeat(43);
async function mount(tok,root){
 dom.window.history.replaceState(null,"","/auth/verify#"+tok);
 await React.act(async()=>{root.render(React.createElement(React.StrictMode,null,React.createElement(VerifyEmailPage,{key:tok})));});
 assert.equal(dom.window.location.hash,"","verification secret must be removed from address bar");
 return dom.window.document.getElementById("root");
}
async function fillPassword(container,value="owner-selected-password-2026"){
 let node=container.querySelector('input[type="password"]');assert.ok(node);
 const setter=Object.getOwnPropertyDescriptor(dom.window.HTMLInputElement.prototype,"value").set;
 await React.act(async()=>{setter.call(node,value);node.dispatchEvent(new dom.window.Event("input",{bubbles:true}));});
}
async function submit(container){
 const form=container.querySelector("form.form-grid");assert.ok(form);
 await React.act(async()=>{form.dispatchEvent(new dom.window.Event("submit",{bubbles:true,cancelable:true}));await Promise.resolve()});
}
function visibleState(container){return {text:container.textContent,ready:!!container.querySelector("a[href='/']"),status:container.querySelector('[role="status"]')?.textContent||""}}
test("independent stale successful response from unmounted token A cannot mark token B ready",async t=>{
 t.after(()=>{globalThis.fetch=originalFetch});
 const pending=defer(),received=[];const A=token("A"),B=token("B");
 globalThis.fetch=async(url,options)=>{let req=JSON.parse(options.body);received.push(req.token);return req.token===A?pending.promise:denied()};
 dom.window.document.body.innerHTML='<div id="root"></div>';
 const container=dom.window.document.getElementById("root"),root=createRoot(container);
 try{
  await mount(A,root);await fillPassword(container);await submit(container);
  assert.deepEqual(received,[A],"negative control: token A request must actually start");
  await mount(B,root);await fillPassword(container);await submit(container);
  assert.deepEqual(received,[A,B],"negative control: new mounted principal's proof must be submitted");
  assert.equal(visibleState(container).ready,false);
  pending.resolve(ok());
  await React.act(async()=>{await pending.promise;await Promise.resolve();});
  assert.equal(visibleState(container).ready,false,"stale A success activated B's visible state");
  assert.match(visibleState(container).status,/Verification could not be completed/);
  assert.equal(dom.window.localStorage.length,0);
  t.diagnostic("INDEPENDENT_UNMOUNTED_STALE_VERIFICATION_ISOLATION=PASS");
 }finally{pending.resolve(denied());await React.act(async()=>root.unmount())}
});
test("same-mounted verification page must not submit superseded token after hash-only principal transition",async t=>{
 t.after(()=>{globalThis.fetch=originalFetch});
 const A=token("C"),B=token("D"),received=[];
 globalThis.fetch=async(url,options)=>{received.push(JSON.parse(options.body).token);return denied()};
 dom.window.document.body.innerHTML='<div id="root"></div>';
 const container=dom.window.document.getElementById("root"),root=createRoot(container);
 try{
  await mount(A,root);await fillPassword(container);
  // A hash-only navigation can preserve the same route/component instance.
  // Re-enter a new email-ownership proof and notify the browser without remount.
  // Use the browser's real fragment setter; jsdom dispatches hashchange
  // asynchronously, preserving the current same-route React component.
  await React.act(async()=>{dom.window.location.hash=B;await new Promise(resolve=>setTimeout(resolve,0));});
  await submit(container);
  assert.equal(received.length,1,"negative control: verification submission did not happen");
  assert.notEqual(received[0],A,"superseded proof A was submitted after navigation to B");
  assert.equal(received[0],B,"current proof B was not used");
  assert.equal(dom.window.location.hash,"","current proof must also be cleared from URL");
  t.diagnostic("INDEPENDENT_SAME_MOUNT_VERIFICATION_CONTEXT_SWITCH=PASS");
 }finally{await React.act(async()=>root.unmount())}
});


for(const outcome of ["success","error"]){
 test(`same-mounted stale ${outcome} cannot overwrite a newer proof`,async t=>{
  t.after(()=>{globalThis.fetch=originalFetch});
  const A=token("E"),B=token("F"),pending=defer(),received=[];
  globalThis.fetch=async(url,options)=>{const req=JSON.parse(options.body);received.push(req.token);assert.equal(options.credentials,"omit");return req.token===A?pending.promise:denied()};
  dom.window.document.body.innerHTML='<div id="root"></div>';const container=dom.window.document.getElementById("root"),root=createRoot(container);
  try{
   await mount(A,root);await fillPassword(container);await submit(container);assert.deepEqual(received,[A]);
   await React.act(async()=>{dom.window.location.hash=B;await new Promise(r=>setTimeout(r,0))});
   assert.equal(dom.window.location.hash,"");await submit(container);assert.deepEqual(received,[A,B]);
   const before=visibleState(container);assert.equal(before.ready,false);
   pending.resolve(outcome==="success"?ok():new Response(JSON.stringify({error:{message:"obsolete failure"}}),{status:400,headers:{"Content-Type":"application/json"}}));
   await React.act(async()=>{await pending.promise;await Promise.resolve()});
   assert.deepEqual(visibleState(container),before,"superseded response changed current UI");
   assert.equal(dom.window.localStorage.length,0);assert.equal(dom.window.sessionStorage.length,0);
  }finally{pending.resolve(denied());await React.act(async()=>root.unmount())}
 });
}

test("network failure permits retry and a second link after completion",async t=>{
 t.after(()=>{globalThis.fetch=originalFetch});const received=[];let failure=true;
 globalThis.fetch=async(url,options)=>{received.push(JSON.parse(options.body).token);if(failure){failure=false;throw new Error("synthetic network failure")};return ok()};
 dom.window.document.body.innerHTML='<div id="root"></div>';const container=dom.window.document.getElementById("root"),root=createRoot(container);
 try{
  await mount(token("G"),root);await fillPassword(container);await submit(container);assert.equal(visibleState(container).ready,false);
  assert.equal(container.querySelector("form.form-grid button").disabled,false);
  await submit(container);assert.equal(visibleState(container).ready,true);
  await React.act(async()=>{dom.window.location.hash=token("H");await new Promise(r=>setTimeout(r,0))});
  assert.equal(visibleState(container).ready,false);assert.equal(dom.window.location.hash,"");await fillPassword(container);await submit(container);
  assert.deepEqual(received,[token("G"),token("G"),token("H")]);assert.equal(visibleState(container).ready,true);
  assert.equal(dom.window.localStorage.length,0);assert.equal(dom.window.sessionStorage.length,0);assert.equal(dom.window.document.cookie,"");
 }finally{await React.act(async()=>root.unmount())}
});

test("silent same-route history navigation captures and removes the new proof",async t=>{
 t.after(()=>{globalThis.fetch=originalFetch});const received=[];
 globalThis.fetch=async(url,options)=>{received.push(JSON.parse(options.body).token);return denied()};
 dom.window.document.body.innerHTML='<div id="root"></div>';const container=dom.window.document.getElementById("root"),root=createRoot(container);
 try{
  await mount(token("I"),root);await fillPassword(container);
  await React.act(async()=>{dom.window.history.pushState({syntheticRouterState:true},"","/auth/verify?local=1#"+token("J"));await new Promise(r=>setTimeout(r,300));});
  assert.equal(dom.window.location.hash,"");assert.equal(dom.window.location.search,"?local=1");assert.deepEqual(dom.window.history.state,{syntheticRouterState:true});
  await submit(container);assert.deepEqual(received,[token("J")]);
 }finally{await React.act(async()=>root.unmount())}
});

test("explicit hashchange after same-route replacement captures B before submission",async t=>{
 t.after(()=>{globalThis.fetch=originalFetch});const received=[];
 globalThis.fetch=async(url,options)=>{received.push(JSON.parse(options.body).token);return denied()};
 dom.window.document.body.innerHTML='<div id="root"></div>';const container=dom.window.document.getElementById("root"),root=createRoot(container);
 try{
  await mount(token("K"),root);await fillPassword(container);
  dom.window.history.replaceState(null,"","/auth/verify#"+token("L"));
  await React.act(async()=>{dom.window.dispatchEvent(new dom.window.Event("hashchange"))});
  assert.equal(dom.window.location.hash,"","new fragment must be removed before any submission");assert.deepEqual(received,[]);
  await submit(container);assert.deepEqual(received,[token("L")],"new proof must replace old proof");
 }finally{await React.act(async()=>root.unmount())}
});
