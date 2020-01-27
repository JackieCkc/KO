CREATE OR REPLACE FUNCTION get_events_aggr() 
RETURNS TABLE (
  app_id int,
  app_veiws int,
  app_installs int,
  app_conversion_rate decimal
)
AS $$
DECLARE
    app_id_row record;
    event_type_row record;
    views integer := 0;
    prev_views integer := 0;
    installs integer := 0;
BEGIN
FOR app_id_row IN
	-- get unique app ids
	SELECT e1.app_id
	FROM store_events e1
	GROUP BY e1.app_id
	ORDER BY e1.app_id ASC
LOOP
	FOR event_type_row IN
		-- iterate each events for an app id
		SELECT e2.event_type
		FROM store_events e2
		WHERE e2.app_id = app_id_row.app_id
		ORDER BY e2.event_time_utc ASC
	LOOP
		IF event_type_row.event_type = 'store_app_view' THEN
			-- handle app views
			-- prev_views is used to tracking 
			views = views + 1;
			prev_views = prev_views + 1;
		END IF;
		
		IF event_type_row.event_type = 'store_app_download' THEN
			-- handle installs for early versions
			-- if there is an app view before, count it as an install
			IF prev_views > 0 THEN
				installs = installs + 1;
				prev_views = prev_views - 1;
			END IF;
		END IF;
		
		IF event_type_row.event_type = 'store_app_install' THEN
			-- handle installs for later versions
			installs = installs + 1;
		END IF;
	END LOOP;

	RETURN QUERY
		SELECT
			app_id_row.app_id AS app_id,
	      	views AS app_veiws,
	       	installs app_installs,
	       	CAST(installs AS decimal)/NULLIF(views, 0) AS app_conversion_rate;

   	-- reset variables
   	views = 0;
   	prev_views = 0;
   	installs = 0;
END LOOP;
END; $$ 
 
LANGUAGE 'plpgsql';

select * FROM get_events_aggr();