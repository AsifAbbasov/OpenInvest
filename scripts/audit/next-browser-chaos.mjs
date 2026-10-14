import assert from "node:assert/strict";
import {chromium} from "/tmp/openinvest-audit-playwright/node_modules/playwright/index.mjs";
const browser=await chromium.launch({headless:true,args:["--no-sandbox"]});
let cases=0,failures=0;
async function scenario(label,cb) {
 const context=await browser.newContext(),page=await context.newPage();
 try{await cb(page);console.log("BROWSER_CASE_PASS="+label);cases++}
 catch(e){console.error("BROWSER_CASE_FAILURE="+label+" "+String(e.stack||e));failures++}
 finally{await context.close()}
}
const token=x=>x.repeat(43);
const url="http://127.0.0.1:3000/auth/verify";
await scenario("real_next_load_scrubs_fragment",async page=>{
 await page.goto(url+"?lang=en#"+token("A"));
 await page.getByRole("heading",{name:"Complete email verification"}).waitFor();
 await page.waitForFunction(()=>window.location.hash==="");
 assert.equal(new URL(page.url()).hash,"");
 assert.equal(await page.evaluate(()=>localStorage.length+sessionStorage.length),0);
});
await scenario("hash_navigation_uses_B_not_A",async page=>{
 const sent=[];
 await page.route("**/api/v1/auth/email/verify",async route=>{
  sent.push(JSON.parse(route.request().postData()).token);
  await route.fulfill({status:400,contentType:"application/json",body:JSON.stringify({error:{code:"INVALID_VERIFICATION",message:"Verification could not be completed."}})});
 });
 await page.goto(url+"#"+token("A"));
 await page.locator('input[type="password"]').fill("browser-audit-password-2026");
 await page.evaluate(b=>{window.location.hash=b},token("B"));
 await page.waitForFunction(()=>window.location.hash==="");
 await page.getByRole("button",{name:"Create verified account"}).click();
 await page.waitForTimeout(100);
 assert.deepEqual(sent,[token("B")]);
});
await scenario("silent_pushState_poll_sanitizes_within_1s",async page=>{
 await page.goto(url+"#"+token("C"));
 await page.evaluate(b=>{history.pushState({__NA:true,audit:"preserve"},"", "/auth/verify?mode=chaos#"+b)},token("D"));
 await page.waitForFunction(()=>location.hash==="",{timeout:1100});
 const obj=await page.evaluate(()=>({state:history.state,search:location.search}));
 assert.equal(obj.state.audit,"preserve");assert.equal(obj.search,"?mode=chaos");
});
await scenario("stale_A_success_after_B_remains_B",async page=>{
 let oldRoute;
 const seen=[];
 await page.route("**/api/v1/auth/email/verify",async route=>{
  const tokenValue=JSON.parse(route.request().postData()).token;seen.push(tokenValue);
  if(tokenValue===token("E"))oldRoute=route;
  else await route.fulfill({status:400,contentType:"application/json",body:JSON.stringify({error:{code:"INVALID_VERIFICATION",message:"Invalid current proof"}})});
 });
 await page.goto(url+"#"+token("E"));
 await page.locator('input[type="password"]').fill("browser-audit-password-2026");
 await page.getByRole("button",{name:"Create verified account"}).click();
 await page.waitForFunction(()=>document.querySelector("button")?.disabled===true);
 await page.evaluate(b=>location.hash=b,token("F"));
 await page.waitForFunction(()=>location.hash==="");
 await page.getByRole("button",{name:"Create verified account"}).click();
 await page.waitForFunction(()=>document.querySelector('[role="status"]')?.textContent.includes("Invalid current proof"));
 await oldRoute.fulfill({status:200,contentType:"application/json",body:JSON.stringify({data:{status:"ACCOUNT_READY"},meta:{requestId:"old-inflight"}})});
 await page.waitForTimeout(100);
 assert.equal(await page.locator('a[href="/"]').count(),0);
 assert.match(await page.locator('[role="status"]').innerText(),/Invalid current proof/);
 assert.deepEqual(seen,[token("E"),token("F")]);
});
await browser.close();
console.log("ACTUAL_NEXTJS_BROWSER_RESULTS passed="+cases+" failed="+failures);
process.exitCode=failures?1:0;
