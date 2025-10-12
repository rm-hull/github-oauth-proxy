export type Secret = {
  name: string;
  clientId: string;
  clientSecret: string;
};

export type Config = {
  port: number;
  logLevel: string;
  nodeEnv: string;
  github: Record<string, Secret>;
  cors: {
    allowedOrigins: string[];
  };
};
