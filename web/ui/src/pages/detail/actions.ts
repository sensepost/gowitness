import { ActionFunction, redirect } from "react-router-dom";
import { toast } from "@/hooks/use-toast";
import * as api from "@/lib/api/api";


const deleteAction: ActionFunction = async ({ params }) => {
  const id = params?.id;

  if (!id) throw new Error("id was not set");

  try {
    await api.post('delete', { id: parseInt(id) });
  } catch (err) {
    toast({
      title: "Error",
      description: `Could not delete result: ${err}`,
      variant: "destructive"
    });

    return null;
  }

  toast({
    description: "Result deleted"
  });

  return redirect("/gallery");
};

export { deleteAction };