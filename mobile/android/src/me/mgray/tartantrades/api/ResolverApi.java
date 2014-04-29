package me.mgray.tartantrades.api;

import android.util.Log;
import org.apache.http.HttpResponse;
import org.apache.http.client.ClientProtocolException;
import org.apache.http.client.HttpClient;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.impl.client.DefaultHttpClient;

import java.io.*;
import java.net.URI;
import java.net.URISyntaxException;

public class ResolverApi {
    private static final String TAG = "TartanTrades";

    private final static String RESOLVER_HOST_PORT = "http://mgray.me:8888";

    public static String login(String username, String password) throws IOException {
        Log.v(TAG, String.format("Logging in as %s (pw: %s) ", username, password));
        HttpResponse response = doGet(String.format("%s/login/?username=%s&password=%s", RESOLVER_HOST_PORT, username, password));
        if (response == null) {
            return "";
        } else {
            return convertStreamToString(response.getEntity().getContent());
        }
    }

    private static HttpResponse doGet(String uri) {
        HttpResponse response = null;
        try {
            HttpClient client = new DefaultHttpClient();
            HttpGet request = new HttpGet();
            request.setURI(new URI(uri));
            response = client.execute(request);
        } catch (URISyntaxException e) {
            e.printStackTrace();
        } catch (ClientProtocolException e) {
            e.printStackTrace();
        } catch (IOException e) {
            e.printStackTrace();
        }
        return response;
    }

    private static String convertStreamToString(InputStream inputStream) throws IOException {
        if (inputStream != null) {
            Writer writer = new StringWriter();

            char[] buffer = new char[1024];
            try {
                Reader reader = new BufferedReader(new InputStreamReader(inputStream, "UTF-8"), 1024);
                int n;
                while ((n = reader.read(buffer)) != -1) {
                    writer.write(buffer, 0, n);
                }
            } finally {
                inputStream.close();
            }
            Log.v(TAG, String.format("Read %s", writer.toString()));
            return writer.toString();
        } else {
            return "";
        }
    }

}
