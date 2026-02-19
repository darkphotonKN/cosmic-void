"use client";

import { useState, useRef, useCallback, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/authStore";
import { apiClient } from "@/utils/api";
import { loadStripe } from "@stripe/stripe-js";
import {
  Elements,
  CardElement,
  useStripe,
  useElements,
} from "@stripe/react-stripe-js";

const stripePromise = loadStripe(
  process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY || "",
);

const PLAN = {
  name: "Cosmic Void Pro",
  productId: "prod_TxVD6tpLpq1NFf",
  price: "$10.00",
  interval: "month",
  features: [
    "Unlimited match history",
    "Exclusive Pro skins & cosmetics",
    "Priority matchmaking",
    "Ad-free experience",
    "Early access to new content",
  ],
};

function CheckoutForm({
  clientSecret,
  onSuccess,
  onError,
}: {
  clientSecret: string;
  onSuccess: (msg: string) => void;
  onError: (msg: string) => void;
}) {
  const stripe = useStripe();
  const elements = useElements();
  const [confirming, setConfirming] = useState(false);
  const [polling, setPolling] = useState(false);
  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const startPolling = useCallback(() => {
    setPolling(true);
    let attempts = 0;
    const maxAttempts = 12; // 12 * 5s = 60s

    pollingRef.current = setInterval(async () => {
      attempts++;
      try {
        const res = await apiClient.checkSubscriptionPermission();
        if (res.has_permission) {
          if (pollingRef.current) clearInterval(pollingRef.current);
          setPolling(false);
          onSuccess("Subscription confirmed! You now have Pro access.");
          return;
        }
      } catch {
        // ignore polling errors, keep trying
      }

      if (attempts >= maxAttempts) {
        if (pollingRef.current) clearInterval(pollingRef.current);
        setPolling(false);
        onSuccess(
          "Payment received! Please refresh the page to see your subscription status.",
        );
      }
    }, 5000);
  }, [onSuccess]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!stripe || !elements) return;

    setConfirming(true);
    onError("");

    const cardElement = elements.getElement(CardElement);
    if (!cardElement) return;

    const { error } = await stripe.confirmCardPayment(clientSecret, {
      payment_method: { card: cardElement },
    });

    if (error) {
      onError(error.message || "Payment failed");
      setConfirming(false);
    } else {
      setConfirming(false);
      startPolling();
    }
  };

  return (
    <form onSubmit={handleSubmit} className="sub-payment-form">
      <div className="sub-card-input-wrapper">
        <CardElement
          options={{
            style: {
              base: {
                fontSize: "16px",
                color: "#e2e8f0",
                "::placeholder": { color: "#64748b" },
                iconColor: "#4ecca3",
              },
              invalid: { color: "#ef4444", iconColor: "#ef4444" },
            },
          }}
        />
      </div>
      {polling ? (
        <div className="sub-inline-message sub-info">
          <span className="sub-btn-spinner" /> Confirming subscription...
        </div>
      ) : (
        <button
          type="submit"
          disabled={!stripe || confirming}
          className="sub-btn-primary sub-btn-full"
        >
          {confirming ? (
            <span className="sub-btn-loading">
              <span className="sub-btn-spinner" />
              Confirming...
            </span>
          ) : (
            "Confirm Payment"
          )}
        </button>
      )}
    </form>
  );
}

export default function SubscriptionPage() {
  const router = useRouter();
  const { memberInfo, isAuthenticated } = useAuthStore();

  const [subscribing, setSubscribing] = useState(false);
  const [clientSecret, setClientSecret] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [hasPermission, setHasPermission] = useState(false);
  const [checkingPermission, setCheckingPermission] = useState(true);

  useEffect(() => {
    if (!isAuthenticated) return;
    apiClient
      .checkSubscriptionPermission()
      .then((res) => {
        if (res.has_permission) {
          setHasPermission(true);
        }
      })
      .catch(() => { })
      .finally(() => setCheckingPermission(false));
  }, [isAuthenticated]);

  if (!isAuthenticated || !memberInfo) {
    return (
      <div className="sub-loading">
        <div className="profile-spinner" />
      </div>
    );
  }

  const handleSubscribe = async () => {
    setError("");
    setSuccess("");
    setSubscribing(true);
    try {
      const res = await apiClient.subscribe(PLAN.productId, memberInfo.email);
      const secret = res.result?.client_secret || res.result?.clientSecret;
      if (!secret) {
        setError("No client secret returned from server");
        setSubscribing(false);
        return;
      }
      setClientSecret(secret);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to subscribe");
    } finally {
      setSubscribing(false);
    }
  };

  return (
    <main className="sub-container">
      <div className="sub-bg" />

      <div className="sub-header">
        <h1 className="sub-title">Subscription</h1>
      </div>

      <div className="sub-card">
        <div className="sub-plan-badge">PRO</div>
        <h2 className="sub-plan-name">{PLAN.name}</h2>

        <div className="sub-plan-price">
          <span className="sub-price-amount">{PLAN.price}</span>
          <span className="sub-price-interval">/ {PLAN.interval}</span>
        </div>

        <ul className="sub-features">
          {PLAN.features.map((feature) => (
            <li key={feature} className="sub-feature-item">
              <svg
                className="sub-check-icon"
                viewBox="0 0 24 24"
                fill="currentColor"
              >
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
              </svg>
              {feature}
            </li>
          ))}
        </ul>

        {checkingPermission ? (
          <div
            className="sub-btn-primary sub-btn-full"
            style={{ textAlign: "center", opacity: 0.6 }}
          >
            <span className="sub-btn-loading">
              <span className="sub-btn-spinner" />
              Loading...
            </span>
          </div>
        ) : hasPermission ? (
          <div className="sub-inline-message sub-success">
            You are subscribed to Cosmic Void Pro!
          </div>
        ) : !clientSecret ? (
          <button
            onClick={handleSubscribe}
            disabled={subscribing}
            className="sub-btn-primary sub-btn-full"
          >
            {subscribing ? (
              <span className="sub-btn-loading">
                <span className="sub-btn-spinner" />
                Processing...
              </span>
            ) : (
              "Subscribe Now"
            )}
          </button>
        ) : (
          <Elements
            stripe={stripePromise}
            options={{
              clientSecret,
              appearance: {
                theme: "night",
                variables: {
                  colorPrimary: "#4ecca3",
                  colorBackground: "#0f0f23",
                  colorText: "#e2e8f0",
                  colorDanger: "#ef4444",
                  borderRadius: "8px",
                  fontFamily: "inherit",
                },
                rules: {
                  ".Input": {
                    border: "1px solid rgba(78, 204, 163, 0.3)",
                    backgroundColor: "rgba(15, 15, 35, 0.8)",
                  },
                  ".Input:focus": {
                    border: "1px solid rgba(78, 204, 163, 0.6)",
                    boxShadow: "0 0 8px rgba(78, 204, 163, 0.2)",
                  },
                  ".Label": {
                    color: "#94a3b8",
                  },
                },
              },
            }}
          >
            <CheckoutForm
              clientSecret={clientSecret}
              onSuccess={(msg) => {
                setSuccess(msg);
                setHasPermission(true);
              }}
              onError={setError}
            />
          </Elements>
        )}

        {error && <div className="sub-inline-message sub-error">{error}</div>}
        {success && (
          <div className="sub-inline-message sub-success">{success}</div>
        )}
      </div>
    </main>
  );
}
