import * as api from "@/lib/api/api";
import * as apitypes from "@/lib/api/types";
import { toast } from "@/hooks/use-toast";

const getData = async (
  setLoading: React.Dispatch<React.SetStateAction<boolean>>,
  setList: React.Dispatch<React.SetStateAction<apitypes.list[]>>
) => {
  setLoading(true);
  try {
    const s = await api.get('list');
    setList(s);
  } catch (err) {
    toast({
      title: "API Error",
      variant: "destructive",
      description: `Failed to get list: ${err}`
    });
  } finally {
    setLoading(false);
  }
};

export { getData };
