/// <reference path="../pb_data/types.d.ts" />

// Custom API endpoint for incrementing view count
// This is safer than allowing direct updates via API rules

routerAdd("POST", "/api/prompts/{id}/view", (e) => {
    const id = e.request.pathValue("id");

    if (!id) {
        e.json(400, { error: "Missing prompt ID" });
        return;
    }

    try {
        // Get current prompt
        const collection = $app.findCollectionByNameOrId("prompts");
        const prompt = $app.findRecordById(collection, id);

        if (!prompt) {
            e.json(404, { error: "Prompt not found" });
            return;
        }

        // Increment view count
        const currentCount = prompt.getInt("view_count") || 0;
        prompt.set("view_count", currentCount + 1);

        // Save the record
        $app.save(prompt);

        e.json(200, {
            success: true,
            view_count: currentCount + 1
        });
    } catch (err) {
        console.log("Error incrementing view count:", err);
        e.json(500, { error: "Failed to update view count" });
    }
});

console.log("✅ View count API endpoint registered: POST /api/prompts/{id}/view");
