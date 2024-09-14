import { toast } from "@/hooks/use-toast";
import * as api from "@/lib/api/api";
import * as apitypes from "@/lib/api/types";
import { differenceInMilliseconds, formatDuration, intervalToDuration, } from 'date-fns';

const getData = async (
  setLoading: React.Dispatch<React.SetStateAction<boolean>>,
  setDetail: React.Dispatch<React.SetStateAction<apitypes.detail | undefined>>,
  setWappalyzer: React.Dispatch<React.SetStateAction<apitypes.wappalyzer>>,
  setDuration: React.Dispatch<React.SetStateAction<string>>,
  // args
  id: string | number,
) => {
  setLoading(true);
  try {
    const [detailData, wappalyzerData] = await Promise.all([
      api.get('detail', { id }),
      api.get('wappalyzer')
    ]);
    setDetail(detailData);
    setWappalyzer(wappalyzerData);

    // calculate duration
    if (detailData.network && detailData.network.length > 0) {
      const probedAt = new Date(detailData.probed_at);
      const lastNetworkEntry = new Date(detailData.network[detailData.network.length - 1].time);
      const durationMs = differenceInMilliseconds(lastNetworkEntry, probedAt);
      const durationObj = intervalToDuration({ start: 0, end: durationMs });
      setDuration(formatDuration(durationObj, { format: ['minutes', 'seconds'] }));
    }
  } catch (err) {
    toast({
      title: "API Error",
      variant: "destructive",
      description: `Failed to get detail: ${err}`
    });
  } finally {
    setLoading(false);
  }
};

const deleteResult = async (id: string): Promise<boolean> => {
  try {
    await api.post('delete', { id });
  } catch (error) {
    toast({
      title: "API Error",
      variant: "destructive",
      description: `Failed to delete result: ${error}`
    });

    return false;
  }
  toast({
    description: "Result deleted"
  });

  return true;
};

export { getData, deleteResult };
