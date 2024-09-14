import * as api from "@/lib/api/api";
import * as apitypes from "@/lib/api/types";
import { toast } from "@/hooks/use-toast";

const getData = async (
  setLoading: React.Dispatch<React.SetStateAction<boolean>>,
  setStats: React.Dispatch<React.SetStateAction<apitypes.statistics | undefined>>,
) => {
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

export { getData };
