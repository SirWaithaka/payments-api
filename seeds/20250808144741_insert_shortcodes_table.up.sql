DELETE
FROM public."mpesa_shortcodes";

INSERT INTO public."mpesa_shortcodes" (id, priority, service, type, shortcode, initiator_name, initiator_password, passphrase,
                                       key, secret, callback_url, created_at, updated_at)
VALUES ('018f7e2a-1b3c-7d4e-9f8a-2c5d6e7f8a9b', 1, 'daraja', 'charge', '174379', 'testapi', 'Safaricom123!!',
        'bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919',
        '7nRVPmgCrfIEseRTTmkLDDqAYAKhhS9KWx0AfYLGj9NVE2C2',
        'Cyq7VrtT1vzQAmPQV1zlrC9MZ2n6py6qqaLYzNgFAx6uDG8sTKYSVoCh8sdplZF7',
        'https://webhook.sirwaithaka.space/webhooks/daraja', current_timestamp, null),
       ('018f7e2a-2c4d-7e5f-a1b2-3d6e9f1a2b3c', 1,'daraja', 'payout', '600991', 'testapi', 'Safaricom123!!',
        'bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919',
        '7nRVPmgCrfIEseRTTmkLDDqAYAKhhS9KWx0AfYLGj9NVE2C2',
        'Cyq7VrtT1vzQAmPQV1zlrC9MZ2n6py6qqaLYzNgFAx6uDG8sTKYSVoCh8sdplZF7',
        'https://webhook.sirwaithaka.space/webhooks/daraja', current_timestamp, null),
       ('018f7e2a-3d5e-7f6a-b2c3-4e7f1a2b3c4d', 1,'daraja', 'transfer', '600979', 'testapi', 'Safaricom123!!',
        'bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919',
        '7nRVPmgCrfIEseRTTmkLDDqAYAKhhS9KWx0AfYLGj9NVE2C2',
        'Cyq7VrtT1vzQAmPQV1zlrC9MZ2n6py6qqaLYzNgFAx6uDG8sTKYSVoCh8sdplZF7',
        'https://webhook.sirwaithaka.space/webhooks/daraja', current_timestamp, null),

       --- QUIKK shortcodes
       ('0187cecc-68a0-7900-8bca-12dfb3b978fc', 2, 'quikk', 'charge', '174379', null, null, null,
        '459e9a652a6e6dfd918aeccdf488e9db',
        'd54c2d5868650a926864510cf8f1f616', null, current_timestamp, null),
       ('0187cecc-68a1-7900-916c-05ba965502b0', 2, 'quikk', 'payout', '511382', null, null, null,
        '459e9a652a6e6dfd918aeccdf488e9db',
        'd54c2d5868650a926864510cf8f1f616', null, current_timestamp, null),
       ('0187cecc-68a2-7900-bd65-2a3f8183c0d2', 2, 'quikk', 'transfer', '174379', null, null, null,
        '459e9a652a6e6dfd918aeccdf488e9db',
        'd54c2d5868650a926864510cf8f1f616', null, current_timestamp, null);
