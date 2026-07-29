import { useState } from "react";
import { replayFetch } from "./openreplay";

type Step = "cart" | "payment" | "failed";

export default function App() {
  const [step, setStep] = useState<Step>("cart");
  const [busy, setBusy] = useState(false);

  const pay = async () => {
    setBusy(true);
    try {
      const response = await replayFetch("/demo-api/checkout", { method: "POST" });
      if (!response.ok) {
        console.error("Checkout failed because the payment provider timed out");
        setStep("failed");
      }
    } finally {
      setBusy(false);
    }
  };

  return (
    <main>
      <div className="eyebrow">Seeded incident</div>
      <h1>Replay checkout demo</h1>
      <p className="muted">This journey produces a browser replay and a correlated backend exception.</p>

      {step === "cart" && (
        <section>
          <div className="product">
            <div>
              <strong>Developer keyboard</strong>
              <p className="muted">Quantity 1</p>
            </div>
            <strong>$129</strong>
          </div>
          <button onClick={() => setStep("payment")}>Checkout</button>
        </section>
      )}

      {step === "payment" && (
        <section>
          <label>
            Card number
            <input data-openreplay-hidden placeholder="Ignored by replay privacy settings" />
          </label>
          <button disabled={busy} onClick={pay}>{busy ? "Processing…" : "Pay now"}</button>
        </section>
      )}

      {step === "failed" && (
        <section className="failure">
          <strong>Payment failed</strong>
          <p>The provider timed out. A correlated error was sent to the logging platform.</p>
          <button onClick={() => setStep("cart")}>Try again</button>
        </section>
      )}
    </main>
  );
}
