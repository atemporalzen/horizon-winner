import {test,expect} from "@playwright/test";
import {readFileSync} from "node:fs";
const amaze=readFileSync(new URL("../../html/amaze.html",import.meta.url),"utf8"),server="http://127.0.0.1:18081",entry="http://entry.rebind.test:18081";
async function setup(page,{method="navigation",payload="AWS Metadata Exfil",behavior="target"}={}){
  const loaded=[];
  const configured=amaze.replace("'xx.xx.xx.xx'","'203.0.113.10'").replace("'dynamic.your.domain'","'dynamic.rebind.test'").replace("'8080'","'18081'").replace("targetPath:          '/'","targetPath:          '/controlled-target'").replace("interval:            '20'","interval:            '1'").replace("runTimeoutSeconds:   180","runTimeoutSeconds:   5").replace("attackMethod:        'navigation'",`attackMethod:        '${method}'`).replace("attackPayload:       'AWS Metadata Exfil'",`attackPayload:       '${payload}'`);
  await page.context().route("http://**.rebind.test:18081/**",async route=>{
    const req=route.request(),u=new URL(req.url());
    if(u.pathname==="/amaze.html")return route.fulfill({status:200,contentType:"text/html",body:configured});
    if(u.pathname==="/controlled-target"){
      if(behavior==="entry-error")return route.fulfill({status:404,headers:{"Content-Type":"text/html","Server-Timing":'horizon-entry;desc="Entry server"'},body:"<html><body>404 page not found</body></html>"});
      if(req.resourceType()==="fetch"&&method==="navigation")return route.abort("blockedbyclient");
      return route.fulfill({status:200,contentType:"text/html",body:"<html><body>Controlled localhost response</body></html>"});
    }
    if(u.pathname.startsWith("/payloads/"))loaded.push(u.pathname);
    const response=await route.fetch({url:server+u.pathname+u.search,headers:{...req.headers(),host:u.host}});return route.fulfill({response});
  });
  await page.goto(entry+"/amaze.html");const frame=page.frameLocator("#frame-0");
  if(method==="navigation")await expect(frame.getByRole("button",{name:"Start navigation test"})).toBeVisible();
  return {frame,loaded};
}
test("Horizon workflow configures popup navigation and AWS payload from amaze only",async({page})=>{
  const {frame,loaded}=await setup(page);const messages=[];page.on("console",x=>messages.push(x.text()));
  const opened=page.waitForEvent("popup");await frame.getByRole("button",{name:"Start navigation test"}).click();const popup=await opened;
  await expect(frame.locator("#rebindingstatus")).toContainText("Readable response via navigation");expect(await popup.locator("body").innerText()).toContain("Controlled localhost response");
  expect(loaded).toEqual(["/payloads/aws-metadata-exfil.js"]);expect(messages.some(x=>x.includes("No collector configured"))).toBeTruthy();
});
test("Horizon fetch remains selectable",async({page})=>{const {frame}=await setup(page,{method:"fetch",payload:"Read Response"});await expect(frame.locator("#rebindingstatus")).toContainText("Readable response via fetch",{timeout:10000});});
test("Horizon selected Hook script has no collision and unrelated payloads are not loaded",async({page})=>{const errors=[];page.on("pageerror",x=>errors.push(x.message));const {loaded}=await setup(page,{payload:"Hook and Control"});expect(loaded).toEqual(["/payloads/hook-and-control.js"]);expect(errors).toEqual([]);});
test("Horizon rejects entry error pages without requiring proof marker",async({page})=>{const {frame}=await setup(page,{behavior:"entry-error"});await frame.getByRole("button",{name:"Start navigation test"}).click();await expect(frame.locator("#rebindingstatus")).toContainText("Timed out",{timeout:10000});});
test("Horizon popup blocking is reported",async({page})=>{const {frame}=await setup(page);await page.frames().find(x=>x.url().includes("soopayload")).evaluate(()=>{window.open=()=>null;});await frame.getByRole("button",{name:"Start navigation test"}).click();await expect(frame.locator("#rebindingstatus")).toContainText("Popup blocked");});
test("Horizon Stop cancels navigation",async({page})=>{const {frame}=await setup(page,{behavior:"entry-error"});await frame.getByRole("button",{name:"Start navigation test"}).click();await frame.getByRole("button",{name:"Stop",exact:true}).click();await expect(frame.locator("#rebindingstatus")).toHaveText("Stopped");});
