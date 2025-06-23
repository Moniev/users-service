from prometheus_client import Counter, Gauge, Histogram


REGISTERED_USERS_TOTAL = Counter(
    "registered_users_total",
    "Total number of registered users"
)
REGISTERED_USERS_TOTAL.inc(0) 


FAILED_LOGINS_TOTAL = Counter(
    "failed_logins_total",
    "Total number of failed login attempts"
)
FAILED_LOGINS_TOTAL.inc(0)


USERS_TOTAL = Gauge(
    "users_total",
    "Total number of users in the database"
)

PASSWORD_HASHING_DURATION_SECONDS = Histogram(
    "password_hashing_duration_seconds",
    "Time spent hashing user passwords"
)