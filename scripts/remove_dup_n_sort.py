import locale
import re

locale.setlocale(locale.LC_ALL, 'tr_TR.utf8')

PRODUCTS_WITHOUT_BRAND_PATTERN = r'(?i)^(?:\d+(?:\.\d+)?\s+(?:kg|gr|cc|ml|Derece|adet|lt|g)|\d+x\d+)\b'

with open('all.csv', 'r', encoding="utf-8") as f:
    lines = [line.replace('"', '') for line in f.readlines()]

lines = [line for line in lines if not re.match(PRODUCTS_WITHOUT_BRAND_PATTERN, line)]
lines = sorted(set(lines), key=locale.strxfrm)

with open('unique_product_names.csv', 'w', encoding="utf-8") as f:
    f.writelines(lines)
