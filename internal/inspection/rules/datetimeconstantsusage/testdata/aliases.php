<?php

// E3: declarations of the legacy alias keep its format.
const ISO8601 = DATE_ISO8601;

class Formats
{
    public const ISO8601 = DATE_ISO8601;
    public const LEGACY_iso8601 = \DateTime::ISO8601;
    public const STAMP = <error descr="DATE_ISO8601 is not ISO-8601 compliant; use DATE_ATOM.">DATE_ISO8601</error>;
    public const PREFIXED = 'Y' . <error descr="DATE_ISO8601 is not ISO-8601 compliant; use DATE_ATOM.">DATE_ISO8601</error>;
}
