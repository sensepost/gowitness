import { WideSkeleton } from "@/components/loading";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle
} from "@/components/ui/card";
import { toast } from "@/hooks/use-toast";
import {
  DatabaseIcon,
  FileTextIcon,
  HardDriveIcon,
  NetworkIcon,
  TerminalIcon
} from "lucide-react";
import { Bar, BarChart, CartesianGrid, XAxis } from "recharts";
import { ChartContainer, ChartLegend, ChartLegendContent, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/components/ui/chart";

import * as api from "@/lib/api/api";
import * as apitypes from "@/lib/api/types";
import {
  useEffect,
  useState
} from "react";

const chartConfig = {
  count: {
    label: "Total Status Code Count",
    color: "hsl(var(--chart-5))",
  },
  code: {
    label: "HTTP Status Code",
    color: "hsl(var(--chart-1))",
  },
} satisfies ChartConfig;

const DashboardPage = () => {
  const [stats, setStats] = useState<apitypes.statistics>();
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    const getData = async () => {
      setLoading(true);
      try {
        const s = await api.get('statistics');
        setStats(s);
      } catch (err) {
        toast({
          title: "API Error",
          variant: "destructive",
          description: `Failed to get statistics: ${err}`
        });
      } finally {
        setLoading(false);
      }
    };
    getData();
  }, []);

  if (loading) return <WideSkeleton />;

  return (
    <div className="space-y-8">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card className="overflow-hidden transition-all hover:shadow-lg">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Database Size
            </CardTitle>
            <DatabaseIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {`${stats ? (stats.dbsize / (1024 * 1024)).toFixed(1) : 0} MB`}
            </div>
          </CardContent>
        </Card>

        <Card className="overflow-hidden transition-all hover:shadow-lg">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Total Results
            </CardTitle>
            <FileTextIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats ? stats.results : 0}
            </div>
          </CardContent>
        </Card>

        <Card className="overflow-hidden transition-all hover:shadow-lg">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Headers
            </CardTitle>
            <HardDriveIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats ? stats.headers : 0}
            </div>
          </CardContent>
        </Card>

        <Card className="overflow-hidden transition-all hover:shadow-lg">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Network Logs
            </CardTitle>
            <NetworkIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats ? stats.networklogs : 0}
            </div>
          </CardContent>
        </Card>

        <Card className="overflow-hidden transition-all hover:shadow-lg">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Console Logs
            </CardTitle>
            <TerminalIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats ? stats.consolelogs : 0}
            </div>
          </CardContent>
        </Card>
      </div>
      <ChartContainer config={chartConfig} className="aspect-auto h-[250px] w-full">
        <BarChart accessibilityLayer data={stats?.response_code_stats}>
          <CartesianGrid vertical={false} />
          <XAxis
            dataKey="code"
            tickLine={false}
            tickMargin={10}
            axisLine={false}
          />
          <ChartTooltip content={<ChartTooltipContent hideLabel />} />
          <ChartLegend content={<ChartLegendContent />} />
          <Bar dataKey="count" fill="var(--color-count)" radius={4} />
        </BarChart>
      </ChartContainer>
    </div>
  );
};

export default DashboardPage;