CREATE TABLE public.msg
(
   cid bigint, 
   ts integer, 
   sub text,
   src text,
   loc bigint,
   ttl integer
) 
WITH (
  OIDS = FALSE
)
;
ALTER TABLE public.msg
  OWNER TO root;
COMMENT ON COLUMN public.msg.cid IS 'combined identifier of a stream (contract and channel hash)';
COMMENT ON COLUMN public.msg.ts IS 'timestamp of the message, in seconds from the year 2000';
COMMENT ON COLUMN public.msg.sub IS 'subchannel part, without the root';
COMMENT ON COLUMN public.msg.src IS 'source filename of the log file';
COMMENT ON COLUMN public.msg.loc IS 'location within the log file (offset and length)';
COMMENT ON COLUMN public.msg.ttl IS 'the expiration time of the value';

CREATE INDEX idx_query ON public.msg (cid, ts, sub text_pattern_ops);
CREATE INDEX idx_ttl ON public.msg (ttl);

-- ALTER TABLE public.msg CLUSTER ON idx_query;


CREATE TABLE public.msg_bucket
(
    oid text,
    ttl integer
)
WITH (
  OIDS = FALSE
)
;

ALTER TABLE public.msg
  OWNER TO root;

CREATE INDEX idx_s3_oid ON public.msg_bucket (oid);
CREATE INDEX idx_s3_ttl ON public.msg_bucket (ttl);