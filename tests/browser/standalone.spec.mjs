import {test,expect} from "@playwright/test";
import {readFileSync} from "node:fs";
const source=readFileSync(new URL("../../amaze.html",import.meta.url),"utf8");
const origin="http://session.rebind.test:8080";
async function setup(page,{behavior="navigation",configured=true,path="/proof",host="session.rebind.test"}={}){
  const html=source.replace("configured: false",`configured: ${configured}`).replace('rebindingHost: ""',`rebindingHost: ${JSON.stringify(host)}`).replace('targetPath: "/proof"',`targetPath: ${JSON.stringify(path)}`).replace("waitBeforeNavigationSeconds: 60","waitBeforeNavigationSeconds: 1").replace("timeoutSeconds: 180","timeoutSeconds: 6");
  await page.context().route(`${origin}/**`,async route=>{
    const r=route.request(),u=new URL(r.url());
    if(u.pathname==="/proof"){
      if(r.resourceType()==="fetch"&&behavior!=="fetch-readable")return route.abort("blockedbyclient");
      const text=behavior==="host-rejected"?"horizon-host-denied":behavior==="wrong-marker"?"Unrelated page":"horizon-lab-proof-v1";
      return route.fulfill({status:200,contentType:"text/html",body:`<!doctype html><html><body>${text}</body></html>`});
    }
    return route.fulfill({status:200,contentType:"text/html",body:html});
  });
  await page.goto(`${origin}/amaze.html`);return html;
}
test("standalone requires explicit configuration",async({page})=>{await setup(page,{configured:false});await expect(page.locator("#start")).toBeDisabled();await expect(page.locator("#setup")).toContainText("Edit CONFIG");});
test("standalone reads proof through a visible popup on plain HTTP",async({page})=>{
  await setup(page);await expect(page.locator("#start")).toBeEnabled();const opened=page.waitForEvent("popup");await page.locator("#start").click();const popup=await opened;
  await expect(page.locator("#state")).toHaveText("navigation after fetch rejection");await expect(page.locator("#response")).toContainText("horizon-lab-proof-v1");expect(await popup.evaluate(()=>window.name)).toMatch(/^horizon-popup-/);
  const download=page.waitForEvent("download");await page.locator("#download").click();expect((await download).suggestedFilename()).toBe("horizon-winner-observation.json");
});
test("standalone reports Host protection",async({page})=>{await setup(page,{behavior:"host-rejected"});await page.locator("#start").click();await expect(page.locator("#state")).toHaveText("host header protected");});
test("standalone does not claim bypass when fetch works",async({page})=>{await setup(page,{behavior:"fetch-readable"});await page.locator("#start").click();await expect(page.locator("#state")).toHaveText("navigation confirmed");await expect(page.locator("#events")).toContainText("fetch_target_readable");});
test("standalone Stop cancels",async({page})=>{await setup(page,{behavior:"wrong-marker"});await page.locator("#start").click();await page.locator("#stop").click();await expect(page.locator("#state")).toHaveText("stopped");});
test("standalone rejects cross-origin target paths",async({page})=>{await setup(page,{path:"//external.invalid/proof"});await expect(page.locator("#state")).toHaveText("Configuration needed");await expect(page.locator("#start")).toBeDisabled();});
test("standalone builds the Singularity hex hostname",async({page})=>{await setup(page,{host:""});await expect(page.locator("#entry")).toHaveAttribute("href",/^http:\/\/s-cb00710a\.7f000001-[a-f0-9]{32}-fs-e\.dynamic\.example\.com:8080\/amaze\.html\?/);});
test("standalone entry HTML marker cannot be mistaken for target proof",async({page})=>{await setup(page,{path:"/amaze.html"});await page.locator("#start").click();await expect(page.locator("#state")).toHaveText("timed out",{timeout:10000});await expect(page.locator("#response")).toHaveText("No target document confirmed yet.");});
