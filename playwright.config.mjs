import {defineConfig} from "@playwright/test";

export default defineConfig({
  testDir:"./tests/browser",fullyParallel:false,workers:1,timeout:20000,
  reporter:"list",use:{headless:true,trace:"retain-on-failure"},
  projects:[{name:"chrome",use:{browserName:"chromium",channel:"chrome"}},{name:"firefox",use:{browserName:"firefox"}}],
  webServer:[{command:"go run ./cmd/horizon-winner -config configs/test.json",url:"http://127.0.0.1:19090/healthz",reuseExistingServer:false,env:{HORIZON_ADMIN_TOKEN:"browser-test-token-01234567890123456789"},timeout:60000},
    {command:"./singularity-server -IPAddress 127.0.0.1 -DNSServerBindAddr 127.0.0.1 -DNSServerPort 15354 -HTTPServerPort 18081 -WsHttpProxyServerPort 13129",url:"http://127.0.0.1:18081/",reuseExistingServer:false,timeout:60000}]
});
