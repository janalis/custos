<?php
function configs(string $dir, array $cases, $runner)
{
    foreach ($cases as $case) {
        file_put_contents($dir . '/config.yaml', $case);
        check($runner->run(['bump']));
        check(Runner::status($dir));
        check(new Report($dir));
        $path = $dir . '/out';
    }
    foreach ($cases as $case) {
        \putenv("MODE=$case");
        app_main($dir);
    }
    foreach ($cases as $case) {
        handle($case);
        <weak_warning descr="Statement does not depend on the loop; move it out.">app_main($dir);</weak_warning>
    }
}
