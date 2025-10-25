from playwright.sync_api import sync_playwright

def test_voting():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()
        try:
            print("Navigating to page...")
            page.goto("http://localhost:8080/page/test-page")
            print("Page loaded.")

            print("Checking initial vote count...")
            vote_count = page.locator(".vote-count")
            print(f"Initial vote count: {vote_count.inner_text()}")
            assert vote_count.inner_text() == "0"
            print("Initial vote count is 0.")

            print("Clicking upvote button...")
            upvote_button = page.locator(".vote-btn", has_text="▲")
            upvote_button.click()
            print("Upvote button clicked.")

            print("Checking updated vote count...")
            page.wait_for_selector(".vote-count:has-text('1')")
            updated_vote_count = page.locator(".vote-count")
            print(f"Updated vote count: {updated_vote_count.inner_text()}")
            assert updated_vote_count.inner_text() == "1"
            print("Updated vote count is 1.")

            print("Taking screenshot...")
            page.screenshot(path="jules-scratch/verification/verification.png")
            print("Screenshot taken.")

        finally:
            browser.close()

if __name__ == "__main__":
    test_voting()
