CREATE TABLE store_events (
   user_id INTEGER NOT NULL,
   app_id INTEGER NOT NULL,
   event_type VARCHAR (64) NOT NULL,
   event_time_utc TIMESTAMP NOT NULL
);

INSERT INTO "store_events"
("user_id","app_id","event_type","event_time_utc")
VALUES
-- app 1
-- test store_open, store_app_update, store_fetch_manifest are not processed
-- expect {app_id: 1, app_views: 0, app_installs: 0, app_conversion_rate: NULL}
(1,1,'store_open','2020-01-24 20:08:08.323242'),
(1,1,'store_app_updat','2020-01-24 20:11:01.215674'),
(1,1,'store_fetch_manifest','2020-01-24 20:11:01.215674'),

-- app 2
-- test store_app_view
-- expect {app_id: 2, app_views: 1, app_installs: 0, app_conversion_rate: 0}
(1,2,'store_app_view','2020-01-24 20:11:01.215674'),

-- app 3
-- test store_app_install
-- expect {app_id: 3, app_views: 0, app_installs: 1, app_conversion_rate: NULL}
(1,3,'store_app_install','2020-01-24 20:12:01.753111'),

-- app 4
-- test store_app_view + store_app_download
-- expect {app_id: 4, app_views: 1, app_installs: 1, app_conversion_rate: 1}
(1,4,'store_app_view','2020-01-24 20:11:01.215674'),
(1,4,'store_app_download','2020-01-24 20:12:01.753111'),

-- app 5
-- test store_app_download before store_app_view
-- expect {app_id: 5, app_views: 1, app_installs: 0, app_conversion_rate: 0}
(1,5,'store_app_download','2020-01-27 17:46:01.329535'),
(1,5,'store_app_view','2020-01-27 17:47:01.329535'),

-- app 6
-- test store_app_view + store_app_download + store_app_install
-- expect {app_id: 6, app_views: 1, app_installs: 2, app_conversion_rate: 2}
(1,6,'store_app_view','2020-01-27 17:47:19.499385'),
(1,6,'store_app_download','2020-01-27 17:48:00.524156'),
(1,6,'store_app_install','2020-01-27 17:48:00.524156'),

-- app 7
-- test store_app_view * 4 + store_app_download + store_app_install
-- expect {app_id: 7, app_views: 4, app_installs: 2, app_conversion_rate: 0.5}
(1,7,'store_app_view','2020-01-27 17:48:29.290573'),
(1,7,'store_app_view','2020-01-27 17:48:29.290573'),
(1,7,'store_app_view','2020-01-27 17:48:29.290573'),
(1,7,'store_app_view','2020-01-27 17:48:29.290573'),
(1,7,'store_app_download','2020-01-27 17:49:00.83342'),
(1,7,'store_app_install','2020-01-27 17:49:00.83342'),

-- app 8
-- test store_app_view + store_app_download for different users
-- expect {app_id: 8, app_views: 1, app_installs: 0, app_conversion_rate: 0}
(1,8,'store_app_view','2020-01-27 17:48:29.290573'),
(2,8,'store_app_download','2020-01-27 17:49:29.290573'),

-- app 9
-- test all combinations for different users
-- expect {app_id: 9, app_views: 2, app_installs: 3, app_conversion_rate: 1.5}
(1,9,'store_app_view','2020-01-27 17:48:29.290573'),
(2,9,'store_app_download','2020-01-27 17:49:29.290573'),
(1,9,'store_app_download','2020-01-27 17:50:29.290573'),
(2,9,'store_app_view','2020-01-27 17:51:29.290573'),
(3,9,'store_app_install','2020-01-27 17:52:29.290573'),
(2,9,'store_app_download','2020-01-27 17:53:29.290573'),
(3,9,'store_app_download','2020-01-27 17:54:29.290573'),
(2,9,'store_app_download','2020-01-27 17:55:29.290573');
