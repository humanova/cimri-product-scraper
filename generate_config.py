import requests
from lxml import html
import json
from concurrent.futures import ThreadPoolExecutor

PAGINATION_XPATH = "/html/body/div/div[4]/div[2]/div/main/div/div[2]/div/div[2]/div[1]/ul"

with open('pages.txt', 'r') as f:
    urls = f.read().splitlines()

def get_page_count(url):
    page = {}
    page['url'] = url

    response = requests.get(url)
    tree = html.fromstring(response.content)

    page_count_list = tree.xpath(PAGINATION_XPATH)[0]

    page_count = page_count_list.xpath('./li[last()-1]/a/text()')[0]
    page_count = int(page_count)

    page['page_count'] = page_count
    return page

if __name__ == "__main__":
    with ThreadPoolExecutor() as executor:
        results = executor.map(get_page_count, urls)

    with open('proxies.txt', 'r') as file:
        proxies = file.read().splitlines()

    data = {"proxies": proxies, "pages": list(results)}
    with open('config.json', 'w') as f:
        json.dump(data, f, indent=2)