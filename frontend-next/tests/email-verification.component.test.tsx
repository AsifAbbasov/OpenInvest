// @ts-nocheck
// Uses the same JSDOM/TSX execution convention as existing mounted tests.
import assert from "node:assert/strict";
import test from "node:test";
import { JSDOM } from "jsdom";

const dom = new JSDOM("<!doctype html><html><body></body></html>", { url: "https://localhost/auth/verify" });
for (const [name, value] of Object.entries({window:dom.window,document:dom.window.document,navigator:dom.window.navigator,HTMLElement:dom.window.HTMLElement,Event:dom.window.Event})) {
 Object.defineProperty(globalThis,name,{configurable:true,writable:true,value});
}
Object.defineProperty(globalThis,"IS_REACT_ACT_ENVIRONMENT",{configurable:true,writable:true,value:true});
const React = await import("react");
const {createRoot} = await import("react-dom/client");
const {default:VerifyEmailPage} = await import("../src/app/auth/verify/page.tsx");

test("verification fragment survives StrictMode effect replay while the URL is cleared",async()=>{
 dom.window.history.replaceState(null,"","/auth/verify#"+"A".repeat(43));
 dom.window.document.body.innerHTML='<div id="root"></div>';
 const container=dom.window.document.getElementById("root")!;const root=createRoot(container);
 try{
  await React.act(async()=>{root.render(React.createElement(React.StrictMode,null,React.createElement(VerifyEmailPage)));});
  const button=[...container.querySelectorAll("button")].find(x=>x.textContent==="Create verified account");
  assert.ok(button);assert.equal(button.disabled,false,"effect replay must not erase the in-memory verification secret");
  assert.equal(dom.window.location.hash,"");assert.equal(dom.window.localStorage.length,0);
 }finally{await React.act(async()=>root.unmount());}
});
