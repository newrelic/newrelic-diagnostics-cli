package pknudsen.example.com.foobar;

/**
 * Java packages
 */
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.MalformedURLException;
import java.net.URL;
import java.util.HashMap;
import java.util.Map;
import java.util.Random;

/**
 * Android packages
 */
import android.app.Activity;
import android.content.Context;
import android.content.Intent;
import android.net.Uri;
import android.os.Bundle;
import android.os.StrictMode;
import android.view.Menu;
import android.view.MenuItem;
import android.util.Log;
import android.view.View;
import android.view.animation.AlphaAnimation;
import android.view.animation.Animation;
import android.view.animation.AnimationSet;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.widget.Button;
import android.widget.TextView;

/**
 * New Relic packages
 */
import com.newrelic.agent.android.FeatureFlag;
import com.newrelic.agent.android.harvest.DeviceInformation;
import com.newrelic.agent.android.instrumentation.SkipTrace;
import com.newrelic.agent.android.logging.AgentLog;
import com.newrelic.agent.android.NewRelic;
import com.newrelic.agent.android.Agent;
import com.newrelic.agent.android.instrumentation.MetricCategory;
import com.newrelic.agent.android.instrumentation.Trace;
import com.newrelic.agent.android.util.NetworkFailure;
import com.newrelic.com.google.gson.Gson;
import com.newrelic.com.google.gson.GsonBuilder;

/**
 * Http request clients
 */
import okhttp3.*;
import retrofit2.*;


/**
 * My custom class packages
 */

import pknudsen.example.com.foobar.requests.OkRequests;

public class MainActivity extends Activity {
    private static final boolean DEVELOPER_MODE = false ;
    /****
     Constants
     ****/

    //set our log tag so we can find our logged messages
    static final String logTag = "com.pknudsen.foobar";
    
    /**
     * Set some Animation values to use later on for text
     **/
    Animation in = new AlphaAnimation(0.0f, 1.0f);
    Animation out = new AlphaAnimation(1.0f, 0.0f);
    AnimationSet as = new AnimationSet(true);
    /**
     * end Animations
     **/

    MessUpEverything messedUp = new MessUpEverything();
    Context context;


    public void sendMessage (View view) {
        //Do something when the button is clicked
        Intent intent = new Intent(this, FragmentTabbedActivity.class);
        startActivity(intent);
    }

    public void sendSms (View view) {
        Intent sendIntent = new Intent(Intent.ACTION_VIEW);
        sendIntent.setData(Uri.parse("sms:"));
        sendIntent.putExtra("sms_body", "Message here");
        startActivity(sendIntent);
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        if (DEVELOPER_MODE) {
            StrictMode.setThreadPolicy(new StrictMode.ThreadPolicy.Builder()
                    //.detectDiskReads()
                    //.detectDiskWrites()
                    //.detectNetwork()
                    //.detectAll()// or .detectAll() for all detectable problems
                    .penaltyLog()
                    .build());
            StrictMode.setVmPolicy(new StrictMode.VmPolicy.Builder()
                    //.detectLeakedSqlLiteObjects()
                    .detectLeakedClosableObjects()
                    //.detectAll()
                    .penaltyLog()
                    //.penaltyDeath()
                    .build());
        }
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        //Action text asking the user for input
        final TextView textView2 = (TextView) findViewById(R.id.textView2);

        /***
         * WebView Setup
         */
        WebView webView = (WebView) findViewById(R.id.webView);
        WebSettings webSettings = webView.getSettings();
        JavaScriptInterface jsInterface = new JavaScriptInterface(this);
        webSettings.setJavaScriptEnabled(true);
        webView.addJavascriptInterface(jsInterface, "Android");
        /***
         * End WebView setup
         */

            //Start the New Relic agent with our app token
            NewRelic.withApplicationToken(BuildConfig.NEW_RELIC_TOKEN)
                    //.
                    .withApplicationVersion("Shawn-1.3")
                    //.withHttpResponseBodyCaptureEnabled(false)
                    //.withCrashReportingEnabled(false)
                    //.usingCollectorAddress("staging-mobile-collector.newrelic.com")
                    //.usingCrashCollectorAddress("staging-mobile-crash.newrelic.com")
                    //.withDefaultInteractions(false)
                    .withLogLevel(AgentLog.DEBUG)
                    .with
                    .start(this.getApplication());

        NewRelic.enableFeature(FeatureFlag.NetworkRequests);


        //create some new Event attributes
        boolean attributeSet = NewRelic.setAttribute("My Custom Attribute", "attribute value");

        /***
         * Animation durations
         */
        in.setDuration(1000);
        in.setStartOffset(1000);

        out.setDuration(500);

        as.addAnimation(out);
        as.addAnimation(in);
        /***
         * End Animation durations
         */

        //Create a deviceInfo object so we can pull some pieces out of the agent
        DeviceInformation deviceInfo = Agent.getDeviceInformation();
        Log.i("Device Info: ", deviceInfo.getOsName());
        Log.i("Device Info: ", deviceInfo.getRunTime());
        Log.i("Device Info: ", deviceInfo.getArchitecture());
        Log.i("Device Info: ", deviceInfo.getModel());
        Log.i("Device Info: ", deviceInfo.getOsVersion());


        /***
         * Oncreate Http requests
         */

        // Okhttp Malformed URL and IOException test
        Thread okThread = new Thread(new Runnable() {
            @Override
            public void run() {
                try
                {
                    String stringUrl = "https://test.com";
                    OkHttpClient httpClient = new OkHttpClient();
                    URL url = new URL(stringUrl);

                    Request request = new Request.Builder()
                            .url(url)
                            .build();
                    okhttp3.Call call = httpClient.newCall(request);
                    okhttp3.Response response = call.execute();
                    if (!response.isSuccessful()) {
                        throw new Exception ("Unexpected code :" + response);
                    }
                } catch (MalformedURLException e) {
                    Log.e("Okhttp test", e.toString());
                } catch (IOException e) {
                    e.printStackTrace();
                } catch (Exception e) {
                    e.printStackTrace();
                    e.getMessage();
                }

            }

        });
        okThread.start();

        //testing getInputStream
        Thread httpThread = new Thread(new Runnable() {
            @Override
            public void run() {
                try {
                    URL u = new URL("http://google.com/_badgiffy.gif");

                    HttpURLConnection connection = (HttpURLConnection) u.openConnection();
                    InputStream is = connection.getInputStream();
                    InputStreamReader isr = new InputStreamReader(is);
                    while (isr.ready()) {
                        isr.read();
                    }
                }
                catch (MalformedURLException e) {
                    e.printStackTrace();
                }
                catch (IOException e) {
                    e.printStackTrace();
                }
            }
        });
        httpThread.start();

        /***
         * End Oncreate Http requests
         */

        /***
         * Buttons to send various working and erroring http requests
         */

        //Okhttp request
        final Button okButton = (Button) findViewById(R.id.okButton);
        okButton.setOnClickListener(new View.OnClickListener(){
            @Override
            public void onClick(View view){


                textView2.startAnimation(as);

                OkRequests okRequest = new OkRequests();
                okRequest.okRun("https://arcane-meadow-5051.herokuapp.com/services/time/est");


                textView2.setText("Request sent!");
            }
        });

        //Java CAT API - Foobar http request - apache's HttpUriRequest and DefaultHttpClient
        final Button javaButton = (Button) findViewById(R.id.javaButton);
        javaButton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {

                //Make a request to a .NET WebAPI box to test x-newrelic-id case sensitivity
                OkRequests fooBar = new OkRequests();
                fooBar.okRun("http://arcane-meadow-5051.herokuapp.com/services/time/PST");
                textView2.startAnimation(as);
                textView2.setText("Request sent!");
            }
        });

        //CAT API (node, dotnet, python, php, ruby - Foobar http request - apache's HttpUriRequest and DefaultHttpClient
        final Button variedAPIButton = (Button) findViewById(R.id.variedAPIButton);
        variedAPIButton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                //Make a request to a .NET WebAPI box to test x-newrelic-id case sensitivity
                OkRequests foobar = new OkRequests();
                foobar.okRun("http://ec2-54-200-51-186.us-west-2.compute.amazonaws.com/api/api/products/1");
                textView2.startAnimation(as);
                textView2.setText("Request sent!");
            }
        });

        //Malformed Foobar http request (error) - apache's HttpUriRequest and DefaultHttpClient
        final Button httpErrorButton = (Button) findViewById(R.id.httpErrorButton);
        httpErrorButton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {

                //Send off a broken Http request to api.newrelic
                OkRequests fooBar = new OkRequests();
                fooBar.okRun("https://api.newrelic.com/v2/applications.json");
                textView2.startAnimation(as);
                textView2.setText("Request sent!");
                textView2.append("\n" + "HTTP Error!");
            }
        });

        // You dun gone and broke it
        final Button crashButton = (Button) findViewById(R.id.crashButton);
        crashButton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                NewRelic.crashNow(messedUp.randomMessage());
            }
        });

        final Button activitySwitchButton = (Button) findViewById(R.id.activitySwitchButton);

        /***
         * End buttons
         */

        /****
         * Example New Relic API calls
         */

        //arbitrary HTTP transaction API calls
        //NewRelic.noticeHttpTransaction(url, statusCode, startTimeMs, endTimeMs, bytesSent, bytesReceived);
        NewRelic.noticeHttpTransaction("http://newrelic.com", 200, System.nanoTime(), System.nanoTime(),100 ,100);

        //arbitrary HTTP error API call
        Log.i(logTag, "This is the unknown call");
        NewRelic.noticeHttpTransaction("http://www,newrelic,com", 403, System.nanoTime(), System.nanoTime(), 100, 100 );
        //NewRelic.noticeHttpTransaction("http://www,newrelic,com", "GET", 200, System.nanoTime(), System.currentTimeMillis(), 100, 100);

        //dummy network failure
        //These normally go inside of a catch block, but for our examples here they are outside
        NewRelic.noticeNetworkFailure("http://api.newrelic.com/badurl", System.nanoTime(), System.nanoTime(), NetworkFailure.exceptionToNetworkFailure(new Exception()));

        //record a custom metric with whatever name, category, count (minimum parameters)
        NewRelic.recordMetric("Custom Metric Name","MyCategory",1.0);
        NewRelic.recordMetric("ZombieMetric", "Network", 1.0);


        //rename an in-flight interaction
        NewRelic.setInteractionName(getClass().getName());
        NewRelic.startInteraction(getClass().getName());

        //New Relic custom method trace instrumentation
        class NRFoobar {

            NRFoobar() {
                foobar();
            }
            //Add New Relic instrumentation via Annotation and specify its category
            @Trace(category = MetricCategory.JSON)
            public void foobar () {
                //foobar something in here
                Gson gson = new GsonBuilder().create();
                gson.toJson("Hello", System.out);
                gson.toJson(123, System.out);
            }
        }

        String UrlString = "";
        long startTime = System.nanoTime();
        NRFoobar foo = new NRFoobar();
        foo.foobar();
        long endTime = System.nanoTime();

        //create a HashMap for storing our attributes
        Map<String, Object> attributes = new HashMap<String, Object>();

        String fooName = "Foobar event";
            attributes.put("URL", UrlString);
            attributes.put("Request Time", (endTime - startTime) / 10000000 ); // request time in seconds
            attributes.put("Test Group", "A | B");
        boolean foobarEvent = NewRelic.recordEvent(fooName, attributes);
        if (!foobarEvent) {
            Log.e(logTag, fooName +": recordEvent Failed. Retry once.");
            NewRelic.recordEvent(fooName, attributes);
        }

        // Create custom attributes for MobileRequestError event type
        Map<String, Object> networkAttributes = new HashMap<String, Object>();
        String eventName = "MobileRequestError";
            // create attributes for map
            networkAttributes.put("param0", "&myParam0");
            networkAttributes.put("param1", "&myParam1");
            networkAttributes.put("param2", "&myParam2");

        boolean networkAttributeCheck = NewRelic.recordCustomEvent(eventName, networkAttributes);

        // check success or write out to log
        if (!networkAttributeCheck) {
            Log.e(logTag, eventName +": recordCustomEvent Failed. Retry once.");
            NewRelic.recordCustomEvent(eventName, networkAttributes);
        }


        /***
         * End New Relic API calls
         */

    }

    @Override
    public boolean onCreateOptionsMenu(Menu menu) {
        // Inflate the menu; this adds items to the action bar if it is present.
        getMenuInflater().inflate(R.menu.menu_main, menu);
        return true;
    }

    @Override
    public boolean onOptionsItemSelected(MenuItem item) {
        // Handle action bar item clicks here. The action bar will
        // automatically handle clicks on the Home/Up button, so long
        // as you specify a parent activity in AndroidManifest.xml.
        int id = item.getItemId();

        //noinspection SimplifiableIfStatement
        if (id == R.id.action_settings) {
            return true;
        }

        return super.onOptionsItemSelected(item);
    }

    private Request getRequest (String url, int method) {
        Request req = new Request.Builder()
                .url(url)
                .build();
        return req;
    }

    private String doSessionWithBody(String url, RequestBody data, int method) throws IOException {
        OkHttpClient mClient = new OkHttpClient();
        Request request = getRequest(url, method);
        okhttp3.Response response = mClient.newCall(request).execute();

        int status = response.code();
        if (!response.isSuccessful()) {
            throw new IOException("Http response code is: " + status + "\n" + response.body().string());
        }
        String responseBody = response.body().string();
        return responseBody;
    }

    public class MessUpEverything {

        @Trace
        public String breakAllTheThings (){

            String broken = this.randomMessage();

            return broken;
        }

        @Trace
        public String randomMessage (){
            String message;

            Random ran = new Random();
            ran.setSeed(1337);

            switch(ran.nextInt(10)) {
                case 0: message = "Ha";
                    break;
                case 1: message = "Haha";
                    break;
                case 2: message = "HaHaHa";
                    break;
                case 3: message = "HaHaHaHa";
                    break;
                case 4: message = "HaHaHaHaHa";
                    break;
                case 5: message = "HaHaHaHaHaHa";
                    break;
                case 6: message = "HaHaHaHaHaHaHa";
                    break;
                case 7: message = "HaHaHaHaHaHaHaHa";
                    break;
                case 8: message = "HaHaHaHaHaHaHaHaHa";
                    break;
                case 9: message = "HaHaHaHaHaHaHaHaHaHa";
                    break;
                default: message = "default Ha";
                    break;
            }
            return message;
        }
    }
}
