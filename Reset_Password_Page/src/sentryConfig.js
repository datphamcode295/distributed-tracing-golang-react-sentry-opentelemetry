import * as Sentry from "@sentry/react";

Sentry.init({
  dsn: process.env.REACT_APP_SENTRY_DSN,
  environment: process.env.REACT_APP_ENVIRONMENT,
  enableTracing: true,
  tracesSampleRate: 1.0,
  debug: true,
  integrations: [],
  instrumenter: "otel",
});