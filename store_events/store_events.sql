CREATE OR REPLACE FUNCTION get_events_aggr() 
RETURNS TABLE (
  app_id int,
  app_views int,
  app_installs int,
  app_conversion_rate decimal
) AS $$
DECLARE
    app_user_row record;
    event_type_row record;
    app_views integer := 0;
    user_views integer := 0;
    installs integer := 0;
    curr_app_id integer := -1;
    prev_app_id integer := -1;
    curr_user_id integer := -1;
    prev_user_id integer := -1;
BEGIN
FOR app_user_row IN
	-- get unique (app id, user id)
	SELECT e1.app_id, e1.user_id
	FROM store_events e1
	GROUP BY e1.app_id, e1.user_id
	ORDER BY e1.app_id ASC, e1.user_id ASC
LOOP
	FOR event_type_row IN
		-- iterate each events for every (app id, user_id)
		SELECT e2.event_type
		FROM store_events e2
		WHERE
			e2.app_id = app_user_row.app_id AND
			e2.user_id = app_user_row.user_id
		ORDER BY e2.event_time_utc ASC
	LOOP
		curr_app_id = app_user_row.app_id;
		curr_user_id = app_user_row.user_id;

		IF curr_app_id != prev_app_id AND prev_app_id != -1 THEN
			-- return an aggregation row for the previous app
		   	RETURN QUERY
				SELECT
					prev_app_id AS app_id,
			      	app_views AS app_views,
			       	installs app_installs,
			       	CAST(installs AS decimal)/NULLIF(app_views, 0) AS app_conversion_rate;
			
			-- reset count for different events
			app_views = 0;
		   	installs = 0;
		   	user_views = 0;
		END IF;

		if curr_user_id != prev_user_id THEN
			-- reset user views for a different user
		   	user_views = 0;
		END IF;

		IF event_type_row.event_type = 'store_app_view' THEN
			-- handle app views
			app_views = app_views + 1;
			user_views = user_views + 1;
		END IF;
		
		IF event_type_row.event_type = 'store_app_download' THEN
			-- handle installs for early versions
			-- if there is an app view before, count it as an install
			IF user_views > 0 THEN
				installs = installs + 1;
				user_views = user_views - 1;
			END IF;
		END IF;
		
		IF event_type_row.event_type = 'store_app_install' THEN
			-- handle installs for later versions
			installs = installs + 1;
		END IF;

		prev_app_id = app_user_row.app_id;
		prev_user_id = app_user_row.user_id;
	END LOOP;
END LOOP;
	-- return an aggregation row for the last app
	RETURN QUERY
		SELECT
			prev_app_id AS app_id,
	      	app_views AS app_views,
	       	installs AS app_installs,
	       	CAST(installs AS decimal)/NULLIF(app_views, 0) AS app_conversion_rate;
END;
$$ LANGUAGE 'plpgsql';

select * from get_events_aggr();