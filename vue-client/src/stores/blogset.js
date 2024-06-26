import { defineStore } from 'pinia'
import axios from 'axios'
import { reactive } from 'vue'

export const useBlogSetStore = defineStore('blogset', () => {
    const data = reactive({
        sitename: "一个神秘站点",
        sitecreatetime: "",
        sitebackground: ""
    })

    async function GetBlogInfo() {
        const response = await axios.get('/queryblogset');
        if (response.data.code === 200) {
            if (response.data.blogsetinfo.sitename !== "") data.sitename = response.data.blogsetinfo.sitename;
            data.sitecreatetime = response.data.blogsetinfo.sitecreatetime;
            data.sitebackground = response.data.blogsetinfo.sitebackground;
        } else {
            console.log(response.data.message);
        }
    }
    return {GetBlogInfo, data}
})