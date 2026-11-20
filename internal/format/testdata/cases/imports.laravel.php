<?php
// presets: per symfony laravel; set: import_order=alpha, single_import_per_statement=true
namespace A {
    use const B\C;

    use A\First;
    use Z\Last;

    use function D\e;

    use F\{G, H};

    new First;
}
