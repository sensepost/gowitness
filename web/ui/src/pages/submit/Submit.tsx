import { useState } from 'react';
import { Form, useNavigation } from 'react-router-dom';
import { PlusCircle, Trash2, Send, Settings } from 'lucide-react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";

export default function JobSubmissionPage() {
  const [urls, setUrls] = useState<string[]>(['']);
  const [advancedOptions, setAdvancedOptions] = useState(false);
  const navigation = useNavigation();

  const handleUrlChange = (index: number, value: string) => {
    const newUrls = [...urls];
    newUrls[index] = value;
    setUrls(newUrls);
  };

  const addUrl = () => {
    setUrls([...urls, '']);
  };

  const removeUrl = (index: number) => {
    const newUrls = urls.filter((_, i) => i !== index);
    setUrls(newUrls);
  };

  return (
    <div className="container mx-auto py-6">
      <Card>
        <CardHeader>
          <CardTitle>Launch a New Probe</CardTitle>
          <CardDescription>Enter URLs and set options for your probe</CardDescription>
        </CardHeader>
        <CardContent>
          <Form method="post" className="space-y-6">
            <div className="space-y-4">
              <h3 className="text-lg font-semibold">URLs</h3>
              {urls.map((url, index) => (
                <div key={index} className="flex items-center space-x-2">
                  <Input
                    type="url"
                    name={`url-${index}`}
                    placeholder="https://sensepost.com"
                    value={url}
                    onChange={(e) => handleUrlChange(index, e.target.value)}
                    className="flex-grow"
                  />
                  {index === urls.length - 1 ? (
                    <Button type="button" variant="outline" size="icon" onClick={addUrl}>
                      <PlusCircle className="h-4 w-4" />
                    </Button>
                  ) : (
                    <Button type="button" variant="outline" size="icon" onClick={() => removeUrl(index)}>
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              ))}
            </div>

            <Accordion type="single" collapsible>
              <AccordionItem value="options">
                <AccordionTrigger>
                  <div className="flex items-center">
                    <Settings className="mr-2 h-4 w-4" />
                    Probe Options
                  </div>
                </AccordionTrigger>
                <AccordionContent>
                  <div className="space-y-4 pt-4">
                    <div className="grid gap-4 md:grid-cols-2">
                      <div className="space-y-2">
                        <Label htmlFor="format">Screenshot Format</Label>
                        <Select name="format" defaultValue="jpeg">
                          <SelectTrigger id="format">
                            <SelectValue placeholder="Select format" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="png">PNG</SelectItem>
                            <SelectItem value="jpeg">JPEG</SelectItem>
                            <SelectItem value="pdf">PDF</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="timeout">Timeout (seconds)</Label>
                        <Input
                          id="timeout"
                          name="timeout"
                          type="number"
                          min="0"
                          defaultValue="60"
                        />
                      </div>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Switch
                        id="advanced-options"
                        checked={advancedOptions}
                        onCheckedChange={setAdvancedOptions}
                      />
                      <Label htmlFor="advanced-options">Advanced Options</Label>
                    </div>

                    {advancedOptions && (
                      <div className="space-y-4">
                        <div className="space-y-2">
                          <Label htmlFor="user-agent">User Agent</Label>
                          <Input
                            id="user-agent"
                            name="user_agent"
                            defaultValue="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"
                          />
                        </div>

                        <div className="grid gap-4 sm:grid-cols-2">
                          <div className="space-y-2">
                            <Label htmlFor="window-x">Window Width</Label>
                            <Input
                              id="window-x"
                              name="window_x"
                              type="number"
                              min="0"
                              defaultValue="1920"
                            />
                          </div>
                          <div className="space-y-2">
                            <Label htmlFor="window-y">Window Height</Label>
                            <Input
                              id="window-y"
                              name="window_y"
                              type="number"
                              min="0"
                              defaultValue="1080"
                            />
                          </div>
                        </div>
                      </div>
                    )}
                  </div>
                </AccordionContent>
              </AccordionItem>
            </Accordion>

            {/* add defaults as hidden inputs because the form will only subit the rendered dom */}
            {/* expanding the accordion will remove these */}
            {!advancedOptions && (
              <>
                <input type="hidden" name="format" value="jpeg" />
                <input type="hidden" name="timeout" value="60" />
                <input type="hidden" name="user_agent" value="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36" />
                <input type="hidden" name="user_agent" value="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36" />
                <input type="hidden" name="window_x" value="1920" />
                <input type="hidden" name="window_y" value="1080" />
              </>
            )}

            <div className="flex justify-end">
              <Button type="submit" disabled={navigation.state === "submitting"}>
                <Send className="mr-2 h-4 w-4" />
                {navigation.state === "submitting" ? "Submitting..." : `Submit ${urls.length} Target${urls.length > 1 ? "s" : ""}`}
              </Button>
            </div>
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}