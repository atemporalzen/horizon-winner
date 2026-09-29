import {test,expect} from "@playwright/test";

const admin="http://127.0.0.1:19090",token="browser-test-token-01234567890123456789";
async function setup(page,request,behavior){
  const r=await request.post(`${admin}/api/sessions`,{headers:{Authorization:`Bearer ${token}`},data:{target_id:"lab"}});expect(r.status()).toBe(201);const session=await r.json();
  // Transport simulation, not a real DNS/LNA test. Keep the browser's URL and
  // origin unchanged while feeding it entry/fixture responses deterministically.
  await page.context().route(`http://${new URL(session.url).host}/**`,async route=>{
    const req=route.request(),u=new URL(req.url());
    if(u.pathname==="/proof"){
      if(req.resourceType()==="fetch"&&behavior!=="fetch-readable")return route.abort("blockedbyclient");
      const body=behavior==="host-rejected"?"horizon-host-denied: unexpected Host":behavior==="wrong-marker"?"unrelated page":"horizon-lab-proof-v1";
      return route.fulfill({status:behavior==="host-rejected"?421:200,contentType:"text/html",body:`<!doctype html><html><body><pre>${body}</pre></body></html>`});
    }
    const upstream=await route.fetch({url:`http://127.0.0.1:18080${u.pathname}${u.search}`,headers:{...req.headers(),host:u.host}});
    await route.fulfill({response:upstream});
  });
  await page.goto(session.url);await expect(page.locator("#start")).toBeEnabled();return session;
}
test("user gesture opens popup, reads proof and downloads report",async({page,request})=>{
  await setup(page,request,"navigation");const popupPromise=page.waitForEvent("popup");await page.locator("#start").click();const popup=await popupPromise;
  await expect(page.locator("#state")).toHaveText("navigation after fetch rejection");
  await expect(page.locator("#response")).toContainText("horizon-lab-proof-v1");
  const downloadPromise=page.waitForEvent("download");await page.locator("#download").click();const download=await downloadPromise;expect(download.suggestedFilename()).toMatch(/^horizon-[a-f0-9]+\.json$/);
  expect(await popup.evaluate(()=>document.body.textContent)).toContain("horizon-lab-proof-v1");
});
test("Host rejection is reported as protection",async({page,request})=>{await setup(page,request,"host-rejected");await page.locator("#start").click();await expect(page.locator("#state")).toHaveText("host header protected");});
test("popup blocker keeps Start available",async({page,request})=>{await setup(page,request,"navigation");await page.evaluate(()=>{window.open=()=>null;});await page.locator("#start").click();await expect(page.locator("#state")).toHaveText("Popup blocked");await expect(page.locator("#start")).toBeEnabled();});
test("Stop cancels the active run",async({page,request})=>{await setup(page,request,"wrong-marker");await page.locator("#start").click();await page.locator("#stop").click();await expect(page.locator("#state")).toHaveText("stopped");await expect(page.locator("#download")).toBeEnabled();});
test("wrong marker never becomes a successful navigation",async({page,request})=>{await setup(page,request,"wrong-marker");await page.locator("#start").click();await expect(page.locator("#state")).toHaveText("timed out",{timeout:12000});await expect(page.locator("#response")).toHaveText("No target document confirmed yet.");});
test("fetch-readable result avoids claiming a blocked-fetch bypass",async({page,request})=>{await setup(page,request,"fetch-readable");await page.locator("#start").click();await expect(page.locator("#state")).toHaveText("navigation confirmed");await expect(page.locator("#events")).toContainText("fetch_target_readable");});
test("controller authenticates and creates a session",async({page})=>{await page.goto(admin);await page.locator("#token").fill(token);await page.locator("#connect button").click();await expect(page.locator("#experiment")).toBeVisible();await page.locator("#create button").click();await expect(page.locator("#run-link")).toHaveAttribute("href",/^http:\/\/s-7f000001\.7f000002-/);});
