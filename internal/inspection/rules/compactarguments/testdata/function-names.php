<?php
namespace Views {
    function compact(...$names) { return $names; }

    function local($title)
    {
        // user function declared in the namespace: names are not variables
        return compact('title', 'subtitle');
    }

    function builtin($title)
    {
        return \Compact('title', <error descr="Variable '$subtitle' may be undefined when compact() runs.">'subtitle'</error>);
    }
}

namespace Pages {
    use function Views\compact;

    function imported($head)
    {
        return compact('head', 'foot');
    }

    function qualified($head)
    {
        return Helpers\compact('head', 'foot');
    }
}

namespace {
    function upper($rows)
    {
        return COMPACT('rows', <error descr="Variable '$total' may be undefined when compact() runs.">'total'</error>);
    }
}
