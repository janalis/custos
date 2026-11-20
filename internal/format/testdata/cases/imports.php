<?php // presets: per symfony laravel; set: import_order=alpha, single_import_per_statement=true
namespace A {
    use \Z\Last, \A\First;
    use const B\C;
    use function D\e;
    use F\{G, H};

    new First;
}
