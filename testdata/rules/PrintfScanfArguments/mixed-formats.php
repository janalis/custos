<?php
final class Notice
{
    private const SHORT = 'Missing %s.';
    private $template = 'Field %s is %s.';

    public static function make(string $message, string $field, bool $verbose, string $detail): string
    {
        // Possible formats mixing a literal with something else: silent.
        $a = sprintf($message ?: 'Expected a value. Got: %s', $field, $detail);
        $b = sprintf($verbose ? 'Field %s missing.' : <<<'TXT'
            Field %s missing: %s
            TXT, $field, $detail);
        $c = sprintf($message ?? 'Missing %s.', $field, $detail);
        $d = sprintf($verbose ? $message : 'Missing %s.', $field, $detail);
        $e = sprintf(self::OTHER ?? 'Missing %s.', $field, $detail);

        // Every format known, but they disagree: silent.
        $f = sprintf($verbose ? 'Field %s: %s' : 'Field %s', $field, $detail);
        $fmt = 'Field %s';
        if ($verbose) {
            $fmt = 'Field %s: %s';
        }
        $g = sprintf($fmt, $field, $detail);

        // A parameter default is only one of the possible formats: silent.
        return $a . $b . $c . $d . $e . $f . $g . self::pad($field);
    }

    private static function pad(string $field, string $format = '[%s] %s'): string
    {
        return sprintf($format, $field);
    }

    public function known(bool $verbose, string $field): string
    {
        // Every format known and they agree: reported.
        $h = <error descr="This call needs 3 argument(s) in total.">sprintf</error>($verbose ? 'Field %s: %s' : '%s => %s', $field);
        $fmt = self::SHORT;
        if ($verbose) {
            $fmt = 'Absent: %s.';
        }
        $i = <error descr="This call needs 2 argument(s) in total.">sprintf</error>($fmt, $field, $verbose);
        $j = <error descr="This call needs 3 argument(s) in total.">sprintf</error>($this->template, $field);
        $k = sprintf(<error descr="Malformed format string.">$verbose ? '100% %s' : '50% %s'</error>, $field);
        return $h . $i . $j . $k;
    }
}
