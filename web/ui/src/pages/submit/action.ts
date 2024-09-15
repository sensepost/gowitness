import { toast } from "@/hooks/use-toast";
import * as api from "@/lib/api/api";
import { redirect } from "react-router-dom";

const submitAction = async ({ request }: { request: Request; }) => {
  const formData = await request.formData();

  // grab submitted urls
  const urls = Array.from(formData.entries())
    .filter(([key]) => key.startsWith('url-'))
    .map(([, value]) => value as string)
    .filter(url => url.trim() !== '');

  if (urls.length === 0) {
    return { error: "Please enter at least one URL" };
  }

  const options = {
    format: formData.get('format'),
    timeout: parseInt(formData.get('timeout') as string),
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
    description: "Probe has been submitted"
  });

  return redirect("/submit");
};

export { submitAction };