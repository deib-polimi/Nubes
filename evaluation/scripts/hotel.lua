require "socket"
local JSON = require("JSON")
math.randomseed(socket.gettime() * 1000)
math.random();
math.random();
math.random()

local user_login_path = ""
local search_path = ""
local recommend_rate_path = ""
local recommend_location_path = ""
local reserve_path = "" -- custom Export for Reservations

-- according to counts specified in the seeder
local max_user_suffix = 49
local max_city_suffix = 4
local max_hotel_suffix = 19
local max_room_suffix = 4
local email_prefix = "Email_"
local city_prefix = "Milano"
local hotel_prefix = "Bruschetti"

local function login()
    local id = math.random(0, max_user_suffix)
    local method = "GET"
    local param = {
        Id = email_prefix .. tostring(id),
        Parameter = "Password" .. tostring(id)
    }
    local body = JSON:encode(param)
    local headers = {}
    headers["Content-Type"] = "application/json"

    return wrk.format(method, user_login_path, headers, body)
end

local function recommend()
    local recommend_rate = math.random(0, 1)
    local city_id = math.random(0, 4)
    local method = "GET"
    local path = ""
    local param = {}

    if recommend_rate == 0 then
        path = recommend_rate_path
        param = {
            Id = "Milano" .. tostring(city_id),
            Parameter =  6
        }
    else
        path = recommend_location_path
        param = {
            Id = "Milano" .. tostring(city_id),
            Count = 6,
            Longitude = (-1)*math.random(0, 90) + math.random(0, 89) + math.random(),
            Latitude =  (-1)*math.random(0, 180) + math.random(0, 179) + math.random(),
        }
    end
    
    local body = JSON:encode(param)
    local headers = {}
    headers["Content-Type"] = "application/json"

    return wrk.format(method, path, headers, body)
end

local function search_hotel()
    local city_id = math.random(0, max_city_suffix)
    local method = "GET"
    local param = {
        Id = "Milano" .. tostring(city_id)
    }
    local body = JSON:encode(param)
    local headers = {}
    headers["Content-Type"] = "application/json"

    return wrk.format(method, search_path, headers, body)
end

local function get_two_consecutive_days_in_year(year, length)
    local day1 = math.random(1, 28)
    local month1 = math.random(1,12)
    local month2 = month1
    local day2 = day1 + length
    local year2 = year

    -- handle the cases when day2 is out of range of a month
    if day2 > 30 and (month1 == 4 or month1 == 6 or month1 == 9 or month1 == 11) then
        month2 = month2 + 1
        day2 = day2 - 30
        if month2 > 12 then
            month2 = 1
            year2 = year + 1
        end
    elseif day2 > 28 and month1==2 then
        if (year%400==0) or ((year%4==0) and (year%100~=0)) then
            -- leap year
            if day2 > 29 then
                day2 = day2 - 29
                month2 = month2 + 1
                if month2 > 12 then
                    month2 = 1
                    year2 = year + 1
                end
            end
        else
            day2 = day2 - 28
            month2 = month2 + 1
            if month2 > 12 then
                month2 = 1
                year2 = year + 1
            end
        end
    elseif day2 > 31 then
        day2 = day2 - 31
        month2 = month2 + 1
        if month2 > 12 then
            month2 = 1
            year2 = year + 1
        end
    end

    -- add leading zeros
    local day1_str = tostring(day1)
    local day2_str = tostring(day2)
    if day1 < 10 then
        day1_str = "0" .. day1_str 
    end
    if day2 < 10 then
        day2_str = "0" .. day2_str 
    end

    
    local month1_str = tostring(month1)
    local month2_str = tostring(month2)
    if month1 < 10 then
        month1_str = "0" .. month1_str 
    end
    if month2_str < 10 then
        month2_str = "0" .. month2_str 
    end

    return day1_str .. "-" .. month1_str .. tostring(year), day2_str .. "-" .. month2_str .. tostring(year2)
end

local function reserve()
    local email_id = math.random(0, max_user_suffix)
    local city_id = math.random(0, max_city_suffix)
    local hotel_id = math.random(0, max_hotel_suffix)
    local room_id = math.random(0, max_room_suffix)

    local method = "GET"
    local param = {
        TypeName = "Reservation",
        Parameter = {
            User = email_prefix .. tostring(email_id),
            RoomId = city_prefix .. tostring(city_id) .. "_" .. hotel_prefix .. tostring(hotel_id) .. "Room" .. tostring(room_id),
        }
    }

    -- in 50% of cases try to reserve a room in cases
    -- where the room is likely to be fully booked 
    local coin = math.random()
    if coin < 0.5 then
        local date1, date2 = get_two_consecutive_days_in_year(2023, math.random(1,14))
        param.Parameter.DateIn = date1
        param.Parameter.DateOut = date2
    else
        local date1, date2 = get_two_consecutive_days_in_year(2024, math.random(1,14))
        param.Parameter.DateIn = date1
        param.Parameter.DateOut = date2
    end

    local body = JSON:encode(param)
    local headers = {}
    headers["Content-Type"] = "application/json"
    return wrk.format(method, reserve_path, headers, body)
end

request = function ()
    local search_ratio = 0.6
    local recommend_ratio = 0.2
    local login_ratio = 0.05
    --local reserve_ratio = 0.005

    -- local search_ratio = 0.6
    -- local recommend_ratio = 0.39
    -- local login_ratio = 0.005
    --local reserve_ratio = 0.005

    local coin = math.random()
    if coin < search_ratio then
        return search_hotel()
    elseif coin < search_ratio + recommend_ratio then
        return recommend()
    elseif coin < search_ratio + recommend_ratio + login_ratio then
        return login()
    else
        return reserve()
    end
end