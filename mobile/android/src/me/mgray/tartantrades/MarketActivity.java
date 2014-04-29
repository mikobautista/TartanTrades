package me.mgray.tartantrades;

import android.os.Bundle;
import android.support.v4.app.FragmentActivity;

public class MarketActivity extends FragmentActivity {
    private static final String TAG = "TartanTrades";
    private String token;


    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.market);
    }
}
