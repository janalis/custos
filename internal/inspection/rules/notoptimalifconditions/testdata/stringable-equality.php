<?php
class Status { public function __toString(): string { return 'OK'; } }
class Plain {}

function queued($r, $id)
{
    // Status converts to a string: the comparison reads its text.
    if ($r instanceof Status && $r == 'QUEUED') {}
    if ($r instanceof Status && "QUEUED-$id" != $r) {}
    if ($r instanceof Status && <weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$r === 'QUEUED'</weak_warning>) {}
    if ($r instanceof Status && <weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$r == null</weak_warning>) {}
    if ($r instanceof Status && <weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$r == 1</weak_warning>) {}
    if ($r instanceof Plain && <weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$r == 'QUEUED'</weak_warning>) {}
    if ($r instanceof $id && <weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$r == 'QUEUED'</weak_warning>) {}
}
