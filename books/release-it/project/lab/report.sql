-- Standard report: one row per 5 s window per route and tenant class.
SELECT (ts_ms // 5000) * 5                                              AS "window",
       route,
       tenant_class                                                     AS class,
       count(*)                                                         AS req,
       count(*) FILTER (status BETWEEN 200 AND 299 AND err_class IS NULL) AS "2xx",
       count(*) FILTER (status BETWEEN 400 AND 499)                     AS "4xx",
       count(*) FILTER (status >= 500)                                  AS "5xx",
       count(*) FILTER (err_class = 'client_timeout' OR status = 504)   AS timeouts,
       round(quantile_cont(latency_ms, 0.50), 1)                        AS p50_ms,
       round(quantile_cont(latency_ms, 0.99), 1)                        AS p99_ms
FROM runs
GROUP BY ALL
ORDER BY 1, 2, 3;
