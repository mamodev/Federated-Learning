import React from "react";
import { TestData } from "./types";
import { toConfig } from "./config";
import { Button } from "@mui/material";

interface SimulationButtonProps {
  config: TestData;
}

type SimulationState = "idle" | "queued" | "running" | "finished";

export function SimulationButton(props: SimulationButtonProps) {
  const { config } = props;

  const [state, setState] = React.useState<SimulationState>("idle");

  const nameRef = React.useRef<string | null>(null);

  const handleRun = () => {
    setState("queued");
    nameRef.current = config.name;
    fetch("http://localhost:8080/run", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(toConfig(config)),
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error("Failed to run simulation");
        }
        setState("running");
      })
      .catch((error) => {
        console.error(error);
        setState("idle");
      });
  };

  // post run status /info
  React.useEffect(() => {
    if (state == "idle" || state == "finished") return;

    const interval = setInterval(() => {
      fetch("http://localhost:8080/info", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ name: nameRef.current }),
      })
        .then((response) => {
          if (!response.ok) {
            throw new Error("Failed to get simulation status");
          }
          return response.json();
        })
        .then((data) => {
          setState(data.status);
        })
        .catch((error) => {
          console.error(error);
          setState("idle");
        });
    }, 1000);

    return () => clearInterval(interval);
  }, [state]);

  return (
    <Button
      variant="contained"
      color="primary"
      disabled={state !== "idle" && state !== "finished"}
      onClick={handleRun}
    >
      {state === "idle"
        ? "Run simulation"
        : state === "queued"
        ? "Queued"
        : state === "running"
        ? "Running"
        : "Finished"}
    </Button>
  );
}
