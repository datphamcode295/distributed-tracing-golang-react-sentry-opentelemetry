import React, { useState } from 'react';
import { ConsoleSpanExporter, SimpleSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { WebTracerProvider } from '@opentelemetry/sdk-trace-web';
import { FetchInstrumentation } from '@opentelemetry/instrumentation-fetch';
import { ZoneContextManager } from '@opentelemetry/context-zone';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { SentryPropagator, SentrySpanProcessor } from '@sentry/opentelemetry-node';
import { Resource } from '@opentelemetry/resources';
import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions'
import {context, trace} from '@opentelemetry/api';

const getData = (email, spanId, sentryId) => fetch('http://localhost:3000/users/reset-password', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'Sentry-Trace': `${sentryId}-${spanId}-1`,
        'Baggage': `sentry-environment=${process.env.REACT_APP_ENVIRONMENT},sentry-public_key=${process.env.REACT_APP_SENTRY_PUBLIC},sentry-trace_id=${sentryId}`
    },
    body: JSON.stringify({ email }),
});


function ForgotPasswordForm() {
    const [message, setMessage] = useState('');
    const [errorMessage, setErrorMessage] = useState('');


    const provider = new WebTracerProvider({
        resource: new Resource({
            [SemanticResourceAttributes.SERVICE_NAME]: 'reset-password',
        })
    });
    
    provider.addSpanProcessor(new SimpleSpanProcessor(new ConsoleSpanExporter()));
    provider.addSpanProcessor(new SentrySpanProcessor());
    
    provider.register({
        contextManager: new ZoneContextManager(),
        propagator: new SentryPropagator(),
    });
    
    registerInstrumentations({
        instrumentations: [new FetchInstrumentation()],
        tracerProvider: provider
    });

    const webTracerWithZone = provider.getTracer('resetEmailPage');

    async function handleSubmit(event) {
        event.preventDefault();

        const email = event.target.email.value;

        const userAgent = window.navigator.userAgent;

        const singleSpan = webTracerWithZone.startSpan('handleSubmit', { 
            attributes:  {
                            'http.scheme': 'http', 
                            'http.user_agent': userAgent,
                            'http.flavor': '1.1',
                            'http.method': 'POST',
                            'http.route': '/users/reset-password',
                            'http.url': 'http://localhost:3000/users/reset-password',
                            'http.host': 'localhost:3000',
                            'http.target': '/users/reset-password',
                            'http.host_port': '3000',
                        }
            }
        );
        singleSpan.setAttribute('email', email);


        context.with(trace.setSpan(context.active(), singleSpan), () => {
        // get trace ID
        const traceId = trace.getSpan(context.active()).spanContext().traceId;
        const spanId = trace.getSpan(context.active()).spanContext().spanId;
        console.log('traceId', traceId);
        console.log('spanId', spanId);
        
        getData(email, spanId, traceId).then(async (response) => {
            console.log('Call api successed')
            const data = await response.json();
            if (response.ok) {
                setMessage(data.message);
                setErrorMessage('');
                singleSpan.end();
            } else {
                setMessage('');
                setErrorMessage(data.error);
                singleSpan.end();
            }
        }).catch((error) => {
            console.log('Call api failed')
            setMessage('');
            setErrorMessage(error.message);
            singleSpan.end();
        });
        });         
    }

    return (
        <div className="container">
            <h1>Forgot Password</h1>
            <p>Enter your email address to reset your password.</p>
            <form id="forgotPasswordForm" onSubmit={handleSubmit}>
                <input type="email" id="email" placeholder="Enter your email" required />
                <button type="submit">Reset Password</button>
            </form>
            <p id="message" style={{ color: errorMessage ? 'red' : 'green' }}>
                {errorMessage || message}
            </p>
        </div>
    );
}

export default ForgotPasswordForm;
