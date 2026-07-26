import { printXml, readLayout } from "./cups.js";

const name = process.argv[2] || "receipt";
const dest = process.argv[3]; // optional override; otherwise .env PRINTER
const model = process.argv[4]; // optional override; otherwise .env PRINTER_MODEL

const xml = readLayout(name);

await printXml(xml, {
  dest,
  model,
  cut: true,
  feed: 3,
});

console.log(
  `Enviado: layout=${name}` +
    (dest ? ` dest=${dest}` : " dest=.env:PRINTER") +
    (model ? ` model=${model}` : ""),
);
