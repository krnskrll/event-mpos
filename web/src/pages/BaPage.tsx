import { useEffect, useState } from "react";
import Layout from "./Layout";
import { api } from "../api/client";
import type { EventResponse } from "../types";

/**
 * BAs work primarily on the Android device. This page just confirms their login worked and shows
 * the event they're assigned to (the same /api/me/event the Android app uses after login).
 */
export default function BaPage() {
  const [event, setEvent] = useState<EventResponse | null>(null);
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    api
      .get<EventResponse>("/api/me/event")
      .then((res) => setEvent(res.data))
      .catch(() => setMsg("No event is assigned to your account yet. Ask an admin to assign you."));
  }, []);

  return (
    <Layout title="Beauty Advisor">
      <div className="card">
        <h2>Your event</h2>
        {event ? (
          <p>
            <strong>{event.name}</strong> <span className="badge">{event.code}</span>
            <br />
            <span className="muted">
              {event.startDate} → {event.endDate}
            </span>
          </p>
        ) : (
          <p className="muted">{msg ?? "Loading…"}</p>
        )}
        <p className="muted" style={{ marginTop: 16 }}>
          Selling happens on the Sunmi Android device. Sign in there with the same credentials.
        </p>
      </div>
    </Layout>
  );
}
