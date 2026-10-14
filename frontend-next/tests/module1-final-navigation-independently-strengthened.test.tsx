// @ts-nocheck
// Independent bounded navigation probes against the real production React page.
// JSDOM mounted-component evidence, NOT a substitute for actual Next.js E2E.
import assert from "node:assert/strict";
import test from "node:test";
import {JSDOM} from "jsdom";
const dom=new JSDOM("<!doctype html><html><body></body></html>",{url:"https://localhost/auth/verify"});
for(const [name,value] of Object.entries({window:dom.window,document:dom.window.document,navigator:dom.window.navigator,HTMLElement:dom.window.HTMLElement,Event:dom.window.Event,HTMLInputElement:dom.window.HTMLInputElement})){
 Object.defineProperty(globalThis,name,{configurable:true,writable:true,value});
}
Object.defineProperty(globalThis,"IS_REACT_ACT_ENVIRONMENT",{configurable:true,writable:true,value:true});
const React=await import("react");
const {createRoot}=await import("react-dom/client");
const {default:VerifyEmailPage}=await import("../src/app/auth/verify/page.tsx");
const initialFetch=globalThis.fetch;
function response(ok){return new Response(JSON.stringify(ok?{data:{status:"ACCOUNT_READY"}}:{error:{code:"INVALID_VERIFICATION",message:"Verification could not be completed. Request new instructions."}}),{status:ok?200:400,headers:{"Content-Type":"application/json"}})}
function deferred(){let resolve;const promise=new Promise(r=>resolve=r);return {promise,resolve}}
const tok=(letter)=>letter.repeat(43);
async function mount(token){
 dom.window.history.replaceState({persistentNextMetadata:"preserve"},"","/auth/verify?lang=en#"+token);
 dom.window.document.body.innerHTML='<div id="root"></div>';
 const container=dom.window.document.getElementById("root"),root=createRoot(container);
 await React.act(async()=>root.render(React.createElement(React.StrictMode,null,React.createElement(VerifyEmailPage))));
 assert.equal(dom.window.location.hash,"");
 assert.equal(dom.window.location.search,"?lang=en");
 assert.equal(dom.window.history.state.persistentNextMetadata,"preserve");
 return {root,container};
}
async function password(c){
 const el=c.querySelector('input[type="password"]');assert.ok(el);
 const setter=Object.getOwnPropertyDescriptor(dom.window.HTMLInputElement.prototype,"value").set;
 await React.act(async()=>{setter.call(el,"independent-auditor-password-2026");el.dispatchEvent(new dom.window.Event("input",{bubbles:true}));});
}
async function submit(c){const form=c.querySelector("form.form-grid");assert.ok(form);await React.act(async()=>{form.dispatchEvent(new dom.window.Event("submit",{bubbles:true,cancelable:true}));await Promise.resolve()})}
function ready(c){return !!c.querySelector('a[href="/"]')}
function status(c){return c.querySelector('[role="status"]')?.textContent||""}
async function cleanup(root){await React.act(async()=>root.unmount());assert.equal(dom.window.localStorage.length,0);assert.equal(dom.window.sessionStorage.length,0);}
test("silent pushState navigation is found by 250ms poll, preserves Next metadata and scrubs B",async t=>{
 t.after(()=>{globalThis.fetch=initialFetch});
 const A=tok("A"),B=tok("B"),seen=[];
 globalThis.fetch=async(_url,opts)=>{seen.push(JSON.parse(opts.body).token);return response(false)};
 const {root,container}=await mount(A);
 try{
  await password(container);
  const metadata={nextNavigation:"preserve-me",position:3};
  dom.window.history.pushState(metadata,"","/auth/verify?lang=ru#"+B);
  // pushState emits neither popstate nor hashchange; wait only one tick of the
  // advertised fixed 250ms poll, not a runaway soak campaign.
  await React.act(async()=>{await new Promise(r=>setTimeout(r,350))});
  assert.equal(dom.window.location.hash,"","250ms poll failed to scrub fragment");
  assert.equal(dom.window.location.search,"?lang=ru");
  assert.deepEqual(dom.window.history.state,metadata,"Next history metadata overwritten by fragment scrub");
  await submit(container);
  assert.deepEqual(seen,[B],"obsolete A submitted despite silent pushState");
  assert.equal(ready(container),false);
 }finally{await cleanup(root)}
});
test("popstate and replaceState-only navigation each selects latest proof without stale submission",async t=>{
 t.after(()=>{globalThis.fetch=initialFetch});
 const A=tok("J"),B=tok("K"),C=tok("L"),seen=[];
 globalThis.fetch=async(_url,opts)=>{seen.push(JSON.parse(opts.body).token);return response(false)};
 const {root,container}=await mount(A);
 try{
  await password(container);
  dom.window.history.replaceState({next:"stateB"},"","/auth/verify?mode=pop#"+B);
  await React.act(async()=>{dom.window.dispatchEvent(new dom.window.PopStateEvent("popstate"));await Promise.resolve()});
  assert.equal(dom.window.location.hash,"");
  await submit(container);
  dom.window.history.replaceState({next:"stateC"},"","/auth/verify?mode=replace#"+C);
  await React.act(async()=>{await new Promise(r=>setTimeout(r,350))});
  assert.equal(dom.window.location.hash,"");
  await submit(container);
  assert.deepEqual(seen,[B,C]);
  assert.equal(ready(container),false);
 }finally{await cleanup(root)}
});
test("inflight A success cannot mark B ready or clear B's outcome after same-route navigation",async t=>{
 t.after(()=>{globalThis.fetch=initialFetch});
 const A=tok("M"),B=tok("N"),old=deferred(),seen=[];
 globalThis.fetch=async(_url,opts)=>{let proof=JSON.parse(opts.body).token;seen.push(proof);return proof===A?old.promise:response(false)};
 const {root,container}=await mount(A);
 try{
  await password(container);await submit(container);
  assert.deepEqual(seen,[A],"positive control: A request must actually be in flight");
  dom.window.location.hash=B;
  await React.act(async()=>{await new Promise(r=>setTimeout(r,15))});
  assert.equal(dom.window.location.hash,"");
  await submit(container);assert.deepEqual(seen,[A,B]);
  assert.match(status(container),/Verification could not be completed/);
  old.resolve(response(true));
  await React.act(async()=>{await old.promise;await Promise.resolve()});
  assert.equal(ready(container),false,"old A success promoted new principal");
  assert.match(status(container),/Verification could not be completed/,"old A overwrote new B failure");
 }finally{old.resolve(response(false));await cleanup(root)}
});
test("inflight A error cannot overwrite B success after actual hash navigation",async t=>{
 t.after(()=>{globalThis.fetch=initialFetch});
 const A=tok("X"),B=tok("Y"),old=deferred(),seen=[];
 globalThis.fetch=async(_url,opts)=>{let proof=JSON.parse(opts.body).token;seen.push(proof);return proof===A?old.promise:response(true)};
 const {root,container}=await mount(A);
 try{
  await password(container);await submit(container);
  assert.deepEqual(seen,[A]);
  dom.window.location.hash=B;
  await React.act(async()=>{await new Promise(r=>setTimeout(r,15))});
  assert.equal(dom.window.location.hash,"");
  await submit(container);
  assert.deepEqual(seen,[A,B]);assert.equal(ready(container),true);
  const before=status(container);
  old.resolve(response(false));
  await React.act(async()=>{await old.promise;await Promise.resolve()});
  assert.equal(ready(container),true,"stale A error reset B ready");
  assert.equal(status(container),before,"stale A error replaced B success message");
  assert.equal(dom.window.localStorage.length,0);assert.equal(dom.window.sessionStorage.length,0);
 }finally{old.resolve(response(false));await cleanup(root)}
});
test("StrictMode setup/replay and unmount never leave navigation polling timers alive",async()=>{
 const originalSet=dom.window.setInterval.bind(dom.window),originalClear=dom.window.clearInterval.bind(dom.window);
 const active=new Set();
 dom.window.setInterval=(fn,ms)=>{assert.equal(ms,250);let id=originalSet(fn,ms);active.add(id);return id};
 dom.window.clearInterval=(id)=>{active.delete(id);return originalClear(id)};
 try{
  for(let i=0;i<4;i++){
   const {root}=await mount(tok("Q"));
   assert.equal(active.size,1,"StrictMode effect replay leaked an interval");
   await cleanup(root);
   assert.equal(active.size,0,"unmount leaked polling timer");
  }
 }finally{for(const id of active)originalClear(id);dom.window.setInterval=originalSet;dom.window.clearInterval=originalClear}
});
