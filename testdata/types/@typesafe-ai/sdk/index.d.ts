declare module "@typesafe-ai/sdk" {
  export function choice(instructions?: any, criteria?: any): any;
  export function noul(instructions?: any, criteria?: any): any;
  export function score(instructions?: any, criteria?: any): any;
  export class TypeSafeClient {
    constructor(opts?: { apiKey?: string; [key: string]: any });
    systemOne(params: { state: any; questions: Record<string, any>; [key: string]: any }): Promise<{
      answers?: Record<string, any>;
      choices?: Record<string, { choice: string }>;
      nouls?: Record<string, { result?: boolean; noul?: number }>;
      scores?: Record<string, { score: number }>;
    }>;
  }
}
