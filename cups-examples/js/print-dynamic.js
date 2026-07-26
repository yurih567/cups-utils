import { printXml, renderPng, writeTempPng } from "./cups.js";

/**
 * Build a simple receipt XML from order data in JavaScript.
 * In production you would usually load a template file and fill placeholders,
 * or generate XML with a template engine (Handlebars, EJS, etc.).
 */
function buildOrderXml(order) {
  const items = order.items
    .map(
      (item) => `
    <Row justify="between">
      <Text size="11">${escapeXml(item.name)}</Text>
      <Text size="11" align="right">${escapeXml(item.total)}</Text>
    </Row>`,
    )
    .join("");

  return `<?xml version="1.0" encoding="UTF-8"?>
<Page width="80mm" padding="4">
  <Column gap="2" align="center">
    <Text align="center" size="18" weight="bold">${escapeXml(order.company)}</Text>
    <Text align="center" size="12">${escapeXml(order.subtitle)}</Text>
    <Spacer height="4"/>
    <Row justify="between">
      <Text size="12" weight="bold">Data: ${escapeXml(order.datetime)}</Text>
      <Text size="12" weight="bold" align="right">Pedido: ${escapeXml(order.code)}</Text>
    </Row>
    <Divider char="-"/>
    ${items}
    <Divider char="-"/>
    <Row justify="between" fill="·" fill-weight="bold">
      <Text size="14" weight="bold">Total</Text>
      <Text size="14" weight="bold" align="right">${escapeXml(order.total)}</Text>
    </Row>
    <Spacer height="4"/>
    <Text align="center" size="11">NÃO É DOCUMENTO FISCAL</Text>
    <Divider char="="/>
  </Column>
</Page>
`;
}

function escapeXml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&apos;");
}

const order = {
  company: "Grand Chef Burger",
  subtitle: "Pedido gerado em JavaScript",
  datetime: "26/07/2026 10:50",
  code: "JS01",
  items: [
    { name: "X-Burguer", total: "29,90" },
    { name: "Refrigerante Lata", total: "8,00" },
  ],
  total: "37,90",
};

const xml = buildOrderXml(order);
const mode = process.argv[2] || "png";

if (mode === "print") {
  const dest = process.argv[3]; // optional; otherwise .env PRINTER
  await printXml(Buffer.from(xml, "utf8"), { dest });
  console.log(`Pedido dinâmico enviado` + (dest ? ` para ${dest}` : " (PRINTER do .env)"));
} else {
  const png = await renderPng(Buffer.from(xml, "utf8"), { assets: null });
  const out = writeTempPng(png, "cups-dynamic-");
  console.log(`PNG dinâmico: ${out}`);
}
