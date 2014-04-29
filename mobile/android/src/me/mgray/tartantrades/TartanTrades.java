package me.mgray.tartantrades;

import android.app.Activity;
import android.content.Intent;
import android.os.Bundle;
import android.os.StrictMode;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import me.mgray.tartantrades.api.ResolverApi;

import java.io.IOException;

public class TartanTrades extends Activity {
    private static final String TAG = "TartanTrades";
    public static final String TOKEN_EXTRA = "token";


    /**
     * Called when the activity is first created.
     */
    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.main);
        // TODO: Don't be lazy
        StrictMode.ThreadPolicy policy = new StrictMode.ThreadPolicy.Builder().permitAll().build();
        StrictMode.setThreadPolicy(policy);

        final EditText username = (EditText) findViewById(R.id.usernameEditText);
        final EditText password = (EditText) findViewById(R.id.passwordEditText);
        Button loginButton = (Button) findViewById(R.id.loginButton);

        loginButton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                try {
                    String token = ResolverApi.login(username.getText().toString(), password.getText().toString());
                    loginToMarket(token);
                } catch (IOException e) {
                    e.printStackTrace();
                }
            }
        });

    }

    private void loginToMarket(String token) {
        Intent intent = new Intent(this, MarketActivity.class);
        intent.putExtra(TOKEN_EXTRA, token);
        startActivity(intent);
    }
}
