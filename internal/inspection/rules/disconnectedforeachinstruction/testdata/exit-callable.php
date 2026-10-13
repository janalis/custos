<?php
foreach ($rows as $row) {
    <weak_warning descr="Statement does not depend on the loop; move it out.">if</weak_warning> ($stop) { exit(...); }
    <weak_warning descr="Statement does not depend on the loop; move it out.">logTitle($title);</weak_warning>
    useRow($row);
}
foreach ($rows as $row) {
    <weak_warning descr="Statement does not depend on the loop; move it out.">if</weak_warning> ($stop) { die(...); }
    <weak_warning descr="Statement does not depend on the loop; move it out.">logTitle($title);</weak_warning>
    useRow($row);
}
foreach ($rows as $row) {
    if ($row) { exit(1); }
    logTitle($title);
    useRow($row);
}
foreach ($rows as $row) {
    if ($row) { ((exit(...))(1)); }
    logTitle($title);
    useRow($row);
}
