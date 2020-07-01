FROM alpine:3.9
LABEL maintainer Bryan J Rodriguez <bryan.j.rodriguez@intel.com>

RUN set -x \
	&& apk add --no-cache \
		haveged

ENTRYPOINT ["haveged"]
CMD ["-F"]
