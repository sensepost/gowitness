import { toast } from "@/hooks/use-toast";
import * as api from "@/lib/api/api";

const bookmarkResult = async (id: string): Promise<boolean> => {
    try {
        await api.post('bookmark', { id });
    } catch (error) {
        toast({
            title: "API Error",
            variant: "destructive",
            description: `Failed to bookmark result: ${error}`
        });

        return false;
    }
    toast({
        description: "Result bookmark inverted"
    });

    return true;
}

export { bookmarkResult };
