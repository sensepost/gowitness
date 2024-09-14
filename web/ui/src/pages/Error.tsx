import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, } from "@/components/ui/card";
import { CircleXIcon, RefreshCcwIcon } from "lucide-react";
import { isRouteErrorResponse, useNavigate, useRouteError } from "react-router-dom";

const ErrorPage = () => {
  const error = useRouteError();
  const navigate = useNavigate();
  console.error(error);

  let errorMessage = "An unknown error occurred";
  let errorDetails = null;

  if (isRouteErrorResponse(error)) {
    errorMessage = error.statusText || error.data;
    errorDetails = error.data;
  } else if (error instanceof Error) {
    errorMessage = error.message;
    errorDetails = error.stack;
  }

  return (
    <div className="flex items-center justify-center min-h-screen bg-background p-4">
      <Card className="w-full max-w-2xl">
        <CardHeader>
          <CardTitle className="text-3xl flex items-center space-x-2">
            <CircleXIcon className="h-8 w-8 text-destructive" />
            <span>Oops! Something went wrong</span>
          </CardTitle>
          <CardDescription className="text-lg">
            We encountered an error while processing your request
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="text-muted-foreground">
            <strong>Error:</strong> {errorMessage}
          </div>
          {errorDetails && (
            <div className="mt-4">
              <strong className="text-muted-foreground">Details:</strong>
              <pre className="mt-2 whitespace-pre-wrap break-words text-sm bg-muted p-4 rounded-md">
                {errorDetails}
              </pre>
            </div>
          )}
          <div className="flex justify-end space-x-4 mt-6">
            <Button variant="outline" onClick={() => navigate(-1)}>
              Go Back
            </Button>
            <Button onClick={() => window.location.reload()}>
              <RefreshCcwIcon className="mr-2 h-4 w-4" />
              Retry
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default ErrorPage;