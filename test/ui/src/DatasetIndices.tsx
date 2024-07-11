/* eslint-disable @typescript-eslint/no-unused-vars */
import { Restore } from "@mui/icons-material";
import {
  Autocomplete,
  IconButton,
  Slider,
  Stack,
  TextField,
} from "@mui/material";
import { BarChart, LineChart, PieChart } from "@mui/x-charts";
import { Dataset, DatasetConfig, Datasets, DSClientsIndices } from "./datasets";
import {
  computeIndices,
  DistributionFunctionName,
  DistributionFunctions,
  getLoadingSerie,
  LoadingFunctionName,
  LoadingFunctions,
  normalizeDistribution,
} from "./distribution";
import { store, useTestField } from "./store";

type Props = {
  test: string;
  dataset: DatasetConfig;
};

function getBarChartData(
  ds: Dataset,
  indx: DSClientsIndices,
  n_clients: number,
  options?: { stack?: string; area?: boolean }
) {
  const series = [];

  for (let i = 0; i < n_clients; i++) {
    const data = [];

    for (let c = 0; c < ds.getNumClasses(); c++) {
      const indices = indx[i][c];

      // approx to 4 decimal places
      let prc = (indices[1] - indices[0]) / ds.classes[c][1];
      prc = Math.round(prc * 1000) / 1000;
      data.push(prc);
    }

    series.push({
      data,
      label: (loc: string) => (loc === "legend" ? `${i}` : `Client ${i}`),
      ...(options || {}),
    });
  }

  return {
    series,
    categories: Array.from({ length: ds.getNumClasses() }, (_, i) => "" + i),
  };
}

function generatePalette(
  [r1, g1, b1]: [number, number, number],
  [r2, g2, b2]: [number, number, number],
  steps: number
) {
  const dr = (r2 - r1) / steps;
  const dg = (g2 - g1) / steps;
  const db = (b2 - b1) / steps;

  return Array.from({ length: steps }, (_, i) => {
    return `rgb(${Math.floor(r1 + dr * i)}, ${Math.floor(
      g1 + dg * i
    )}, ${Math.floor(b1 + db * i)})`;
  });
}

export function DatasetIndices(props: Props) {
  const { dataset: d, test } = props;

  const {
    distribution_fn,
    distribution_base,
    bias,
    knolwedge_amount,
    loading_fn,
  } = d;

  const rounds = useTestField(test, "rounds") / 2;
  const n_clients = useTestField(test, "n_clients");

  const DS = Datasets[d.name as keyof typeof Datasets];

  const indices = computeIndices(
    DS,
    n_clients,
    distribution_fn,
    bias,
    distribution_base,
    knolwedge_amount
  );

  const barChart = getBarChartData(DS, indices, n_clients);

  const palette = generatePalette([0, 0, 255], [255, 0, 0], n_clients);

  const distrSerie = normalizeDistribution(
    DS,
    distribution_fn,
    bias,
    distribution_base
  );

  const loadingSerie = getLoadingSerie(rounds, loading_fn);

  const clientsDsAmount = Array.from({ length: n_clients }, (_, i) => {
    return Object.keys(indices[i]).reduce((acc, c) => {
      const [start, end] = indices[i][c as unknown as number];
      return acc + (end - start);
    }, 0);
  });

  return (
    <div>
      <Slider
        value={distribution_base}
        onChange={(_, value) =>
          store.changeDatasetField(test, "distribution_base", value as number)
        }
        step={0.01}
        min={0}
        max={1}
      />
      <Slider
        value={bias}
        onChange={(_, value) =>
          store.changeDatasetField(test, "bias", value as number)
        }
        step={0.01}
        min={0}
        max={1}
      />

      <Autocomplete
        disableClearable
        options={Object.keys(DistributionFunctions)}
        value={distribution_fn}
        onChange={(_, value) =>
          store.changeDatasetField(
            test,
            "distribution_fn",
            value as DistributionFunctionName
          )
        }
        renderInput={(params) => <TextField {...params} />}
      />

      <LineChart
        series={[
          {
            data: distrSerie,
          },
        ]}
        height={300}
      />

      <Stack direction="row" spacing={2} alignItems="center">
        <IconButton
          onClick={() =>
            store.changeDatasetField(test, "knolwedge_amount", 1 / n_clients)
          }
        >
          <Restore />
        </IconButton>

        <Slider
          value={knolwedge_amount}
          onChange={(_, value) =>
            store.changeDatasetField(test, "knolwedge_amount", value as number)
          }
          step={0.001}
          min={0.001}
          max={1}
          marks={[{ value: 1 / n_clients, label: "1/n_clients" }]}
        />
      </Stack>

      <Stack direction="row" spacing={2} alignItems="center">
        <BarChart
          series={barChart.series}
          xAxis={[
            {
              scaleType: "band",
              data: barChart.categories,
              tickPlacement: "middle",
              tickLabelPlacement: "middle",

              valueFormatter: (v, ctx) => {
                if (ctx.location === "tick") {
                  return v;
                }

                return `Class ${v}`;
              },
            },
          ]}
          colors={palette}
          slotProps={{ legend: { hidden: true } }}
          height={300}
        />

        <PieChart
          series={[
            {
              data: [
                ...clientsDsAmount.map((v, i) => {
                  return {
                    id: i,
                    value: v,
                    label: `Client ${i}`,
                  };
                }),
                {
                  id: n_clients,
                  value:
                    DS.classes.reduce((acc, [_, v]) => acc + v, 0) -
                    clientsDsAmount.reduce((acc, v) => acc + v, 0),

                  color: "grey",
                  label: "Unassigned",
                },
              ],
            },
          ]}
          width={400}
          height={200}
          colors={palette}
          slotProps={{ legend: { hidden: true } }}
        />
      </Stack>

      <LineChart
        series={barChart.series}
        xAxis={[
          {
            scaleType: "band",
            data: barChart.categories,
            tickPlacement: "middle",
            tickLabelPlacement: "middle",

            valueFormatter: (v, ctx) => {
              if (ctx.location === "tick") {
                return v;
              }

              return `Class ${v}`;
            },
          },
        ]}
        slotProps={{ legend: { hidden: true } }}
        height={300}
        colors={palette}
      />

      <Autocomplete
        disableClearable
        options={Object.keys(LoadingFunctions)}
        value={loading_fn}
        onChange={(_, value) =>
          store.changeDatasetField(
            test,
            "loading_fn",
            value as LoadingFunctionName
          )
        }
        renderInput={(params) => (
          <TextField {...params} label="Round loading function" />
        )}
      />

      <LineChart
        series={[
          {
            data: loadingSerie,
          },
        ]}
        yAxis={[
          {
            min: 0,
            max: 1,
          },
        ]}
        height={300}
      />

      <BarChart
        series={Array.from({ length: n_clients }, (_, i) => {
          const idx = indices[i];

          return {
            label: `Client ${i}`,
            data: loadingSerie.map((v) => {
              return Object.keys(idx).reduce((acc, c) => {
                const [start, end] = idx[c as unknown as number];
                return acc + Math.floor(v * (end - start));
              }, 0);
            }),
          };
        })}
        slotProps={{ legend: { hidden: true } }}
        height={300}
        colors={palette}
      />
    </div>
  );
}
