package me.mgray.tartantrades.mobile.client.ui;

import com.google.gwt.core.client.GWT;
import com.google.gwt.http.client.Request;
import com.google.gwt.http.client.RequestCallback;
import com.google.gwt.http.client.Response;
import com.google.gwt.uibinder.client.UiBinder;
import com.google.gwt.uibinder.client.UiField;
import com.google.gwt.user.client.ui.Composite;
import com.google.gwt.user.client.ui.HTML;
import com.googlecode.mgwt.dom.client.event.tap.TapEvent;
import com.googlecode.mgwt.dom.client.event.tap.TapHandler;
import com.googlecode.mgwt.ui.client.widget.*;
import me.mgray.tartantrades.mobile.client.ResolverAdapter;

public class LoginPanel extends Composite{
    interface LoginPanelUiBinder extends UiBinder<LayoutPanel, LoginPanel> {
    }

    private static LoginPanelUiBinder ourUiBinder = GWT.create(LoginPanelUiBinder.class);
    private MTextBox usernameTextBox = new MTextBox();
    private MPasswordTextBox passwordTextBox = new MPasswordTextBox();

    @UiField(provided = true)
    FormListEntry username = new FormListEntry("Username", usernameTextBox);

    @UiField(provided = true)
    FormListEntry password = new FormListEntry("Password", passwordTextBox);

    @UiField
    Button loginButton;

    @UiField
    HTML token;

    public LoginPanel() {
        LayoutPanel rootElement = ourUiBinder.createAndBindUi(this);
        initWidget(rootElement);
        ResolverAdapter resolverAdapter = new ResolverAdapter();
        resolverAdapter.checkAlive();

        loginButton.addTapHandler(new TapHandler() {
            @Override
            public void onTap(TapEvent tapEvent) {
                System.out.println("Button tapped");
            }
        });
    }
}