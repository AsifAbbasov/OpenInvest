import assert from "node:assert/strict";
import test from "node:test";
const api = await import("../src/common/api/openinvest.ts");
test("registration acknowledgment contains no session and the request carries no password",async(t)=>{
 const previous=globalThis.fetch;t.after(()=>{globalThis.fetch=previous});let observed;
 globalThis.fetch=async(url,init)=>{observed={url,init};return new Response(JSON.stringify({data:{status:"REQUEST_ACCEPTED",message:"If this address is eligible, instructions will be sent."},meta:{}}),{status:202,headers:{"Content-Type":"application/json"}})};
 const result=await api.register({email:"owner@example.test"});assert.equal(result.ok,true);assert.equal(result.data.status,"REQUEST_ACCEPTED");assert.equal(result.data.session,undefined);assert.deepEqual(JSON.parse(observed.init.body),{email:"owner@example.test"});
});
test("verification uses explicit POST without cookies and never returns a session",async(t)=>{
 const previous=globalThis.fetch;t.after(()=>{globalThis.fetch=previous});let observed;
 globalThis.fetch=async(url,init)=>{observed={url,init};return new Response(JSON.stringify({data:{status:"ACCOUNT_READY"},meta:{}}),{status:200,headers:{"Content-Type":"application/json"}})};
 const result=await api.verifyEmailRegistration({token:"synthetic-local-token",password:"synthetic-password",language:"en",theme:"system",timezone:"UTC"});assert.equal(result.ok,true);assert.equal(result.data.session,undefined);assert.equal(observed.init.method,"POST");assert.equal(observed.init.credentials,"omit");assert.ok(observed.url.endsWith("/api/v1/auth/email/verify"));
});
