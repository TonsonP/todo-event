import amqp from "amqplib";
import type { Channel, ConsumeMessage } from "amqplib";

type AmqpConnection = Awaited<ReturnType<typeof amqp.connect>>;

let connection: AmqpConnection | undefined;
let channel: Channel | undefined;

type AuditMessage = {
  type: string;
  payload: unknown;
};

type LokiPush = {
  streams: LokiStream[];
};

type LokiStream = {
  stream: Record<string, string>;
  values: [string, string][];
};

const AMQP_URL = process.env.AMQP_URL || "amqp://guest:guest@localhost:5672/";
const LOKI_URL = process.env.LOKI_URL || "http://localhost:3100";

const TASK_EXCHANGE = "task.events";
const ONBOARDING_EXCHANGE = "onboarding.events";

const QUEUE_AUDIT_TASK_EVENTS = "audit.task.events";
const QUEUE_AUDIT_USER_EVENTS = "audit.user.events";

async function pushToLoki(
  lokiURL: string,
  eventType: string,
  payload: unknown
): Promise<void> {
  const line = JSON.stringify({
    event_type: eventType,
    payload,
  });

  const body: LokiPush = {
    streams: [
      {
        stream: {
          service: "audit",
          event_type: eventType,
        },
        values: [[Date.now().toString() + "000000", line]],
      },
    ],
  };

  const response = await fetch(`${lokiURL}/loki/api/v1/push`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });

  if (!response.ok) {
    throw new Error(`loki push: status ${response.status}`);
  }
}

async function subscribe(
  channel: Channel,
  exchange: string,
  queueName: string,
  handler: (message: AuditMessage) => Promise<void>
): Promise<void> {
  await channel.assertExchange(exchange, "fanout", {
    durable: true,
  });

  await channel.assertQueue(queueName, {
    durable: true,
  });

  await channel.bindQueue(queueName, exchange, "");

  await channel.consume(
    queueName,
    async (rawMessage: ConsumeMessage | null) => {
      if (!rawMessage) {
        return;
      }

      try {
        const parsed = JSON.parse(rawMessage.content.toString()) as AuditMessage;

        if (!parsed.type) {
          throw new Error("message missing type");
        }

        console.info("audit: received event", {
          type: parsed.type,
          exchange,
          queue: queueName,
        });

        await handler(parsed);

        channel.ack(rawMessage);
      } catch (error) {
        console.error("audit: failed to process message", {
          error,
          exchange,
          queue: queueName,
        });

        // false, false = do not requeue
        // Change to false, true if you want retry by requeue.
        channel.nack(rawMessage, false, false);
      }
    },
    {
      noAck: false,
    }
  );
}

async function main(): Promise<void> {
  let connection: AmqpConnection | undefined;
  let channel: Channel | undefined;

  try {
    connection = await amqp.connect(AMQP_URL);
    channel = await connection.createChannel();

    await channel.prefetch(10);

    await subscribe(
      channel,
      TASK_EXCHANGE,
      QUEUE_AUDIT_TASK_EVENTS,
      async (message) => {
        await pushToLoki(LOKI_URL, message.type, message.payload);
      }
    );

    await subscribe(
      channel,
      ONBOARDING_EXCHANGE,
      QUEUE_AUDIT_USER_EVENTS,
      async (message) => {
        await pushToLoki(LOKI_URL, message.type, message.payload);
      }
    );

    console.info("audit service listening", {
      exchanges: [TASK_EXCHANGE, ONBOARDING_EXCHANGE],
    });

    const shutdown = async (signal: NodeJS.Signals) => {
      console.info("audit service stopping", { signal });

      try {
        await channel?.close();
        await connection?.close();
      } catch (error) {
        console.error("audit: shutdown error", { error });
      }

      process.exit(0);
    };

    process.on("SIGINT", shutdown);
    process.on("SIGTERM", shutdown);
  } catch (error) {
    console.error("audit service failed to start", { error });

    try {
      await channel?.close();
      await connection?.close();
    } catch {
      // ignore cleanup error
    }

    process.exit(1);
  }
}

main();