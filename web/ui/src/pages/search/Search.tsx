import { Link, useLoaderData, useNavigation } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import * as api from "@/lib/api/api";
import * as apitypes from "@/lib/api/types";
import { WideSkeleton } from "@/components/loading";
import { getStatusColor } from "@/lib/common";


export default function SearchResultsPage() {
  const data = useLoaderData() as apitypes.searchresult[] | undefined;
  const navigation = useNavigation();

  if (!data || data.length === 0) {
    return <div className="text-center mt-8">No results found.</div>;
  }

  if (navigation.state === 'loading') return <WideSkeleton />;

  return (
    <div className=" mx-auto p-4">
      <h1 className="text-2xl font-bold mb-4">
        Search Results <span className="text-muted-foreground">({data.length})</span>
      </h1>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {data.map((result) => (
          <Link to={`/screenshot/${result.id}`} key={result.id} className="group">
            <Card className="flex flex-col h-full transition-shadow hover:shadow-lg">
              <CardHeader className="relative p-0">
                <img
                  src={
                    result.screenshot
                      ? `data:image/png;base64,${result.screenshot}`
                      : api.endpoints.screenshot.path + "/" + result.file_name}
                  alt={result.url}
                  loading="lazy"
                  className="w-full h-48 object-cover transition-all duration-300 filter group-hover:scale-105"
                />
                <Badge
                  className={`absolute top-2 right-2 ${getStatusColor(result.response_code)} text-white`}
                >
                  {result.response_code}
                </Badge>
              </CardHeader>
              <CardContent className="flex-grow p-4">
                <CardTitle className="text-lg mb-2 line-clamp-2">{result.title}</CardTitle>
                <p className="text-sm text-muted-foreground mb-2 line-clamp-1">{result.final_url}</p>
                <div className="mb-2">
                  <div className="flex flex-wrap gap-2">
                    <p className="text-sm font-semibold mb-1">Matched Fields:</p>
                    {result.matched_fields.map((field) => (
                      <Badge key={field} variant="outline">
                        {field}
                      </Badge>
                    ))}
                  </div>
                </div>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}