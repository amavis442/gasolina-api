// ABOUTME: Tiny Deno utility that reads the device_secret from config.json and serves a QR code page
// ABOUTME: Run: deno run --allow-read --allow-net main.ts [path/to/config.json]

import QRCode from "npm:qrcode@1.5.4";

const configPath = Deno.args[0] ?? "./config.json";

let secret: string;
try {
  const raw = await Deno.readTextFile(configPath);
  secret = JSON.parse(raw).device_secret as string;
  if (!secret) throw new Error("device_secret not found in config");
} catch (e) {
  console.error(`Could not read device_secret from ${configPath}: ${e}`);
  Deno.exit(1);
}

const dataUrl: string = await QRCode.toDataURL(secret, { width: 320, margin: 2 });

const html = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Gasolina — Device Secret QR</title>
  <style>
    body { font-family: sans-serif; display: flex; flex-direction: column;
           align-items: center; justify-content: center; min-height: 100vh;
           margin: 0; background: #f5f5f5; }
    img  { background: white; padding: 20px; border-radius: 8px;
           box-shadow: 0 2px 8px rgba(0,0,0,.15); }
    p    { color: #888; font-size: 13px; margin-top: 12px; word-break: break-all;
           max-width: 340px; text-align: center; }
  </style>
</head>
<body>
  <h2>Scan in Gasolina Settings</h2>
  <img src="${dataUrl}" alt="QR code" />
  <p>${secret}</p>
</body>
</html>`;

const port = 9876;
Deno.serve({ port }, (_req) => new Response(html, {
  headers: { "content-type": "text/html; charset=utf-8" },
}));

console.log(`\nQR code ready → http://localhost:${port}\n`);
console.log(`Secret: ${secret}`);
