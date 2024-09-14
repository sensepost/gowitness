import { toast } from "@/hooks/use-toast";
import * as apitypes from "@/lib/api/types";

const copyToClipboard = (content: string, type: string) => {
  navigator.clipboard.writeText(content).then(() => {
    toast({
      description: `${type} copied to clipboard`,
    });
  }).catch((err) => {
    console.error('Failed to copy content: ', err);
    toast({
      title: "Error",
      description: "Failed to copy content",
      variant: "destructive",
    });
  });
};

const getIconUrl = (tech: string, wappalyzer: apitypes.wappalyzer | undefined): string | undefined => {
  if (!wappalyzer || !(tech in wappalyzer)) return undefined;

  return wappalyzer[tech];
};

const getStatusColor = (code: number) => {
  if (code >= 200 && code < 300) return "bg-green-500 text-white";
  if (code >= 400 && code < 500) return "bg-yellow-500 text-black";
  if (code >= 500) return "bg-red-500 text-white";
  return "bg-gray-500 text-white";
};

export { copyToClipboard, getIconUrl, getStatusColor };