import { Box, Stack, Typography } from "@mui/material";
import { blue, grey, purple } from "@mui/material/colors";
import React, { useEffect, useRef } from "react";
import { REvent, Timeline } from "./types";

export const CELL_SIZE = 30;
export const CCELL_SIZE = 50;

export interface TimelineProps {
  n_clients: number;
  rounds: number;
  timeline: Timeline;
  onChange: (t: Timeline) => void;
}

export function TimelineCmp({
  n_clients,
  timeline,
  onChange,
  rounds,
}: TimelineProps) {
  const timelineRef = useRef<HTMLDivElement>(null);
  const isMouseDownRef = useRef<"" | "right" | "left" | "middle">("");
  // Left click: COMM
  // Right click: RETR
  // Middle click: Reset (NA)

  const [selectedClients, setSelectedClients] = React.useState<number[]>([]);

  useEffect(() => {
    if (!timelineRef.current) return;

    const el = timelineRef.current;

    const handleMouseUp = () => {
      isMouseDownRef.current = "";
    };

    const handleMouseDown = (e: MouseEvent) => {
      const isMiddle = e.button == 1;
      const isLeft = e.button == 0;

      isMouseDownRef.current = isMiddle ? "middle" : isLeft ? "left" : "right";
    };

    const handleContextMenu = (e: MouseEvent) => {
      e.preventDefault();
    };

    el.addEventListener("mousedown", handleMouseDown);
    el.addEventListener("mouseleave", handleMouseUp);
    el.addEventListener("mouseup", handleMouseUp);
    el.addEventListener("contextmenu", handleContextMenu);

    return () => {
      el.removeEventListener("mousedown", handleMouseDown);
      el.removeEventListener("mouseleave", handleMouseUp);
      el.removeEventListener("mouseup", handleMouseUp);
      el.removeEventListener("contextmenu", handleContextMenu);
    };
  }, [timelineRef]);

  const setValue = (t: number, c: number, type: REvent | null) => {
    if (timeline[t] == undefined) timeline[t] = [];

    timeline[t] = timeline[t].filter((x) => x.client != c);
    if (type != null)
      timeline[t].push({
        type: type,
        client: c,
      });

    if (timeline[t].length == 0) delete timeline[t];

    onChange({ ...timeline });
  };

  return (
    <Stack
      ref={timelineRef}
      sx={{
        border: 1,
        width: "fit-content",
        userSelect: "none",
        "& .row:not(:last-child)": {
          borderBottom: "1px solid",
        },

        "& .row": {
          display: "flex",
          flexDirection: "row",
        },

        "& .cell.header": {
          "&.cell-client": {
            backgroundColor: purple[200],
          },
          backgroundColor: purple[200],
        },

        "& .cell": {
          // height: CELL_SIZE,
          "&:not(:last-child)": {
            borderRight: "1px solid",
          },

          "&.cell-client": {
            width: CCELL_SIZE,
          },

          textWrap: "nowrap",
          width: CELL_SIZE,

          display: "flex",
          justifyContent: "center",
          alignItems: "center",
        },
      }}
    >
      <Stack className="row">
        <Stack
          className="cell header cell-client"
          sx={{ width: CCELL_SIZE, textWrap: "nowrap" }}
        >
          C-xxx
        </Stack>
        {Array.from({ length: rounds }, (_, j) => {
          const handleMouseEnter = () => {
            const btn = isMouseDownRef.current;
            if (!btn) return;

            if (selectedClients.length == 0) {
              for (let i = 0; i < n_clients; i++) {
                switch (btn) {
                  case "left":
                    setValue(j, i, "COMM");
                    break;
                  case "right":
                    setValue(j, i, "RETR");
                    break;
                  case "middle":
                    setValue(j, i, null);
                    break;
                }
              }
              return;
            } else {
              for (const i of selectedClients) {
                switch (btn) {
                  case "left":
                    setValue(j, i, "COMM");
                    break;
                  case "right":
                    setValue(j, i, "RETR");
                    break;
                  case "middle":
                    setValue(j, i, null);
                    break;
                }
              }
            }
          };

          return (
            <Stack
              direction="row"
              className="cell header"
              key={j}
              sx={{ width: CELL_SIZE, textWrap: "nowrap" }}
              onMouseEnter={handleMouseEnter}
              onMouseDown={handleMouseEnter}
            >
              {j}
            </Stack>
          );
        })}
      </Stack>

      {Array.from({ length: n_clients }, (_, i) => (
        <TimelineRow
          key={i}
          idx={i}
          timeline={timeline}
          isMouseDownRef={isMouseDownRef}
          rounds={rounds}
          setValue={setValue}
          selected={selectedClients.includes(i)}
          onToggle={() => {
            if (selectedClients.includes(i)) {
              setSelectedClients(selectedClients.filter((x) => x != i));
            } else {
              setSelectedClients([...selectedClients, i]);
            }
          }}
          onPaste={() => {
            if (selectedClients.length == 1) {
              const idx = selectedClients[0];
              for (let t = 0; t < rounds; t++) {
                const t_events = timeline[t] ?? [];
                const evt = t_events.find((x) => x.client == idx);
                const el = evt ?? { type: "NA", client: idx };
                setValue(t, i, el.type == "NA" ? null : el.type);
              }
            }
          }}
          onReset={() => {
            for (let t = 0; t < rounds; t++) {
              setValue(t, i, null);
            }
          }}
        />
      ))}

      <TimelineRow
        idx={-1}
        onPaste={() => {}}
        timeline={timeline}
        isMouseDownRef={isMouseDownRef}
        rounds={rounds}
        setValue={setValue}
        selected={selectedClients.includes(-1)}
        onToggle={() => {
          if (selectedClients.includes(-1)) {
            setSelectedClients(selectedClients.filter((x) => x != -1));
          } else {
            setSelectedClients([...selectedClients, -1]);
          }
        }}
        onReset={() => {
          for (let t = 0; t < rounds; t++) {
            setValue(t, -1, null);
          }
        }}
      />
    </Stack>
  );
}

interface TimelineRowProps {
  idx: number;
  timeline: Timeline;
  isMouseDownRef: React.MutableRefObject<"" | "right" | "left" | "middle">;
  rounds: number;
  setValue: (t: number, c: number, type: REvent | null) => void;
  selected: boolean;
  onToggle: () => void;
  onPaste: () => void;
  onReset: () => void;
}

function TimelineRow(props: TimelineRowProps) {
  const {
    idx,
    timeline,
    isMouseDownRef,
    rounds,
    setValue,
    selected,
    onToggle,
    onPaste,
    onReset,
  } = props;

  return (
    <Stack className="row">
      <Box
        onDoubleClick={onReset}
        onClick={onToggle}
        onContextMenu={(e) => {
          e.preventDefault();
          onPaste();
        }}
        className="cell cell-client"
        sx={{
          backgroundColor: selected ? blue[300] : grey[200],
          cursor: "pointer",
        }}
      >
        {idx >= 0 ? `C-${idx}` : "Agg"}
      </Box>
      {Array.from({ length: rounds }, (_, t) => {
        const t_events = timeline[t] ?? [];

        const evt = t_events.find((x) => x.client == idx);
        const el = evt ?? { type: "NA", client: idx };

        const handleMouseEnter = () => {
          const btn = isMouseDownRef.current;
          if (!btn) return;

          const ovverrideEvent = idx == -1 ? "AGG" : (null as REvent | null);

          switch (btn) {
            case "left":
              setValue(t, idx, ovverrideEvent ?? "COMM");
              break;
            case "right":
              setValue(t, idx, ovverrideEvent ?? "RETR");
              break;
            case "middle":
              setValue(t, idx, null);
              break;
          }
        };

        return (
          <Box
            className="cell"
            key={t}
            onMouseEnter={handleMouseEnter}
            onMouseDown={handleMouseEnter}
          >
            <Typography
              sx={{
                cursor: "pointer",

                color:
                  el.type == "COMM"
                    ? "green"
                    : el.type == "RETR"
                    ? "red"
                    : el.type == "AGG"
                    ? "blue"
                    : grey[200],
              }}
            >
              {el.type == "COMM"
                ? "C"
                : el.type == "RETR"
                ? "R"
                : el.type == "AGG"
                ? "A"
                : "NA"}
            </Typography>
          </Box>
        );
      })}
    </Stack>
  );
}