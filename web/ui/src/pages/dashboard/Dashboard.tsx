import { WideSkeleton } from "@/components/loading";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle
} from "@/components/ui/card";
import { toast } from "@/hooks/use-toast";
import * as api from "@/lib/api/api";
import * as apitypes from "@/lib/api/types";
import {
  useEffect,
  useState
} from "react";

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
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Database Size</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{
            stats
              ? (stats.dbsize / (1024 * 1024)).toFixed(1)
              : 0
          } MB</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Results</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{
            stats
              ? stats.results
              : 0
          }</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Headers</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{
            stats
              ? stats.headers
              : 0
          }</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Network Logs</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{
            stats
              ? stats.networklogs
              : 0
          }</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Console Logs</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{
            stats
              ? stats.consolelogs
              : 0
          }</div>
        </CardContent>
      </Card>
    </div>
  );
};

export default DashboardPage;