package me.mgray.tartantrades.mobile.client;

import com.google.gwt.http.client.*;

public class ResolverAdapter {

    public static final String RESOLVER_URL = "http://www.mgray.me:8888";

    public void getLoginToken(String username, String password) {
        String loginUrl = RESOLVER_URL + "/login/?username=" +username + "&password=" + password;
    }

    public void checkAlive() {
        doGet(RESOLVER_URL);
    }

    public static void doGet(String url) {
        RequestBuilder builder = new RequestBuilder(RequestBuilder.GET, url);

        try {
            Request response = builder.sendRequest(null, new RequestCallback() {
                public void onError(Request request, Throwable exception) {
                    exception.printStackTrace();
                }

                public void onResponseReceived(Request request, Response response) {
                    System.out.println("RESP: " + response.getText() + " --- " + response.getHeadersAsString());
                }
            });

        } catch (RequestException e) {
            // Code omitted for clarity
        }
    }
}
