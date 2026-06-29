import Layout from "./Layout";

/** Placeholder — the real order-fulfilment board lands in EPIC-09. */
export default function WarehousePage() {
  return (
    <Layout title="Warehouse">
      <div className="card">
        <h2>Order fulfilment</h2>
        <p className="muted">
          You are signed in as warehouse staff. The live packing board (NEW → PACKING → READY →
          COLLECTED) arrives in EPIC-09. Authentication and access control are working — that is
          what EPIC-01 delivers.
        </p>
      </div>
    </Layout>
  );
}
