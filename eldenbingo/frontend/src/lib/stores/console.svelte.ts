export type LogLevel = "info" | "warn";

export interface ConsoleLog {
  time: string;
  level: LogLevel;
  message: string;
}

class ConsoleStore {
  logs = $state<ConsoleLog[]>([]);

  add(level: LogLevel, message: string) {
    this.logs.push({
      time: new Date().toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      }),
      level,
      message,
    });

    if (this.logs.length > 500) {
      this.logs.shift();
    }
  }

  clear() {
    this.logs = [];
  }
}

export const consoleStore = new ConsoleStore();
