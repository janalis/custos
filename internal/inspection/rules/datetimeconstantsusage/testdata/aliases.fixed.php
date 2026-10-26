<?php

// E3: declarations of the legacy alias keep its format.
const ISO8601 = DATE_ISO8601;

class Formats
{
    public const ISO8601 = DATE_ISO8601;
    public const LEGACY_iso8601 = \DateTime::ISO8601;
    public const STAMP = DATE_ATOM;
    public const PREFIXED = 'Y' . DATE_ATOM;
}
