<?php
namespace Negative0 { htmlspecialchars($s); }
namespace Negative1 { htmlspecialchars(htmlspecialchars($s),double_encode:false); }
namespace Negative2 { htmlspecialchars(htmlentities($s)); }
namespace Negative3 { htmlspecialchars(htmlspecialchars($s),ENT_QUOTES); }
namespace Negative4 { htmlspecialchars(htmlspecialchars($s,ENT_QUOTES),ENT_NOQUOTES); }
namespace Negative5 { htmlspecialchars(htmlspecialchars($s),double_encode:$flag); }
namespace Negative6 { trim($s); }
