<?php
class Report
{
    public function render($fh, $name, $total, $raw)
    {
        echo sprintf('%05.2f %s', $total, $name);
        echo sprintf("%2\$'#8s %1\$s", $name, $total);
        [$x, $y] = sscanf($raw, '%d-%d');
        $parts = fscanf($fh, '%s %s');
        store(sscanf($raw, '%s'));
        sscanf($raw, '%*d %s', $only);
        sscanf($raw, '%[a-z]', $word);
        printf('%s %s', ...$pair);
        echo sprintf($unknown, $name);
        echo sprintf("Hi $name %s");
        echo sprintf('99% sure', 1);
        echo sprintf('   ');
        $this->sprintf('%s');
        vsprintf('%s %s', [1]);
        $fmt = sprintf(...);
        $n = (sscanf($raw, '%d'));
    }

    public function grow()
    {
        $tpl = 'id=%d';
        $tpl .= ' name=%s';
        return sprintf($tpl, 1, 'x');
    }
}

// The 2-argument form used as a value in any expression.
function scan_values(string $t)
{
    $hms = implode(':', sscanf($t, 'PT%dH%dM%dS') ?? []);
    $n = count(sscanf($t, '%d-%d') ?: []);
    return [$hms, $n, sscanf($t, '%s')];
}

// A redeclarable format property is not resolved (a subclass may change it).
class Dict
{
    public $dropIndex = 'DROP INDEX %s';

    public function drop($idx, $table)
    {
        return sprintf($this->dropIndex, $idx, $table);
    }
}

class MyDict extends Dict
{
    public $dropIndex = 'DROP INDEX %s ON %s';
}
