import { toast } from "@/hooks/use-toast";
import * as api from "@/lib/api/api";
import { parseTargets } from "@/lib/targets";
import { redirect } from "react-router";

const submitJobAction = async ({ formData }: { formData: FormData; }) => {

  // grab submitted urls (newline separated)
  const { urls } = parseTargets((formData.get('urls') as string | null) ?? '');

  if (urls.length === 0) {
    toast({
      title: "Error",
      description: "Please enter at least one valid URL",
      variant: "destructive"
    });
    return null;
  }

  const options = {
    format: formData.get('format'),
    timeout: parseInt(formData.get('timeout') as string),
    delay: parseInt(formData.get('delay') as string),
    user_agent: formData.get('user_agent'),
    window_x: parseInt(formData.get('window_x') as string),
    window_y: parseInt(formData.get('window_y') as string),
  };

  try {
    await api.post('submit', { urls, options });
  } catch (err) {
    toast({
      title: "Error",
      description: `Could not submit new probe: ${err}`,
      variant: "destructive"
    });
    return null;
  }

  toast({
    title: "Success!",
    description: `Probe has been submitted for ${urls.length} target${urls.length === 1 ? "" : "s"}`
  });

  return redirect("/submit");
};

const submitImmediateAction = async ({ formData }: { formData: FormData; }) => {
  const url = formData.get('immediate-url') as string;
  const options = {
    format: formData.get('format'),
    timeout: parseInt(formData.get('timeout') as string),
    delay: parseInt(formData.get('delay') as string),
    user_agent: formData.get('user_agent'),
    window_x: parseInt(formData.get('window_x') as string),
    window_y: parseInt(formData.get('window_y') as string),
  };

  try {
    return await api.post('submitsingle', { url, options });
  } catch (err) {
    toast({
      title: "Error",
      description: `Could not submit new probe: ${err}`,
      variant: "destructive"
    });
    return null;
  }
};

export { submitJobAction, submitImmediateAction };
