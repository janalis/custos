<?php
namespace Negative0 { date('o-W'); }
namespace Negative1 { date('Y-m-d W'); }
namespace Negative2 { date('Y'); }
namespace Negative3 { date('\\Y-W'); }
namespace Negative4 { $x->format('Y-W'); }
namespace MonthOnly { date('F Y, \w\e\e\k W'); date('Y-m W'); date('n/Y W'); date('M Y W'); }
namespace DayOnly { date('Y/z W'); date('Y W d'); date('Y W j'); }
namespace YearFacts { date('L Y W'); date('t Y W'); }
namespace Composite { gmdate('c W Y'); gmdate('r W Y'); gmdate('Y W U'); }
